package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

var deletableTypes = map[string]bool{
	"msg.text": true,
	"msg.file": true,
}

// deleteMessage is called by function "CreateEvent"
// when the event is of type "msg.delete"
func (bs *BoxApplication) deleteMessage(ctx context.Context, receivedEvent events.Event, handler events.EventHandler) (result events.View, err error) {
	toDelete, err := sqlboiler.FindEvent(ctx, bs.db, receivedEvent.ReferrerID.String)
	if err != nil {
		return result, merror.Transform(err).Describe("retrieving event to delete")
	}

	// Authorization-related checks should come as soon as possible
	// so we put them first.
	if receivedEvent.SenderID != toDelete.SenderID {
		// box admins can delete messages even if they are not the author
		if err := events.MustBeAdmin(ctx, bs.db, toDelete.BoxID, receivedEvent.SenderID); err != nil {
			return result, merror.Transform(err).Describe("checking admins")
		}
	}

	isDeleted, err := events.IsDeleted(toDelete)
	if err != nil {
		return result, merror.Transform(err).Describe("checking if event is already deleted")
	}
	if isDeleted {
		return result, merror.Gone().Describe("event is already deleted")
	}

	if !deletableTypes[toDelete.Type] {
		return result, merror.Forbidden().
			Describef("cannot delete events of type \"%s\"", toDelete.Type)
	}

	var fileID string
	if toDelete.Type == "msg.file" {
		// Retrieval of fileID is done *before* we delete the message content
		// but removal of the file is done *after*
		// because file will not be "orphan" before we apply the removal.
		msgFileContent := &events.MsgFileContent{}
		err = json.Unmarshal(toDelete.Content.JSON, &msgFileContent)
		if err != nil {
			return result, merror.Transform(err).Describe("unmarshaling content of event to delete")
		}
		fileID = msgFileContent.EncryptedFileID
	}

	newContentJSON := events.DeletedContent{}
	newContentJSON.Deleted.AtTime = time.Now()
	newContentJSON.Deleted.ByIdentityID = receivedEvent.SenderID

	newContentBytes, err := json.Marshal(newContentJSON)
	if err != nil {
		return result, merror.Transform(err).Describe("marshalling new event content")
	}
	toDelete.Content = null.JSONFrom(newContentBytes)

	tx, err := bs.db.BeginTx(ctx, nil)
	if err != nil {
		return result, merror.Transform(err).Describe("creating DB transaction")
	}

	rowsAff, err := toDelete.Update(ctx, tx, boil.Infer())
	if err != nil {
		atomic.SQLRollback(ctx, tx, err)
		return result, merror.Transform(err).Describe("updating event")
	}
	if rowsAff != 1 {
		atomic.SQLRollback(ctx, tx, err)
		return result, merror.Transform(err).Describef("%d rows affected", rowsAff)
	}

	err = toDelete.Reload(ctx, tx)
	if err != nil {
		atomic.SQLRollback(ctx, tx, err)
		return result, merror.Transform(err).Describe("reloading event")
	}

	// (potential) removal of the actual encrypted file (on S3)
	// is done at the very end because the operation cannot be rolled back
	if fileID != "" {
		isOrphan, err := files.IsOrphan(ctx, tx, fileID)
		if err != nil {
			atomic.SQLRollback(ctx, tx, err)
			return result, merror.Transform(err).Describe("checking if file is orphan")
		}
		if isOrphan {
			if err := files.Delete(ctx, tx, bs.filesRepo, fileID); err != nil {
				atomic.SQLRollback(ctx, tx, err)
				return result, merror.Transform(err).Describe("deleting stored file")
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return result, merror.Transform(err).Describe("committing transaction")
	}

	// not important to wait for after handlers to return
	// NOTE: we construct a new context since the actual one will be destroyed after the function has returned
	subCtx := context.WithValue(ajwt.SetAccesses(context.Background(), ajwt.GetAccesses(ctx)), logger.CtxKey, logger.FromCtx(ctx))
	go func(ctx context.Context, e events.Event) {
		for _, after := range handler.After {
			if err := after(ctx, &receivedEvent, bs.db, bs.redConn, bs.identities); err != nil {
				// we log the error but we don’t return it
				logger.FromCtx(ctx).Warn().Err(err).Msgf("after %s event", receivedEvent.Type)
			}
		}
	}(subCtx, receivedEvent)

	event := events.FromSQLBoiler(toDelete)

	identityMap, err := events.MapSenderIdentities(ctx, []events.Event{event}, bs.identities)
	if err != nil {
		return result, merror.Transform(err).Describe("retrieving identities for view")
	}

	view, err := events.FormatEvent(event, identityMap)
	if err != nil {
		return view, merror.Transform(err).Describe("computing event view")
	}

	return view, nil
}
