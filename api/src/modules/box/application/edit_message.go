package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

// editMessage is called by function "CreateEvent"
// when the event is of type "msg.edit"
func (bs *BoxApplication) editMessage(ctx context.Context, receivedEvent events.Event, handler events.EventHandler) (result events.View, err error) {
	var content events.MsgEditContent
	err = json.Unmarshal(receivedEvent.JSONContent, &content)
	if err != nil {
		return result, merror.Internal().Describe("unmarshaling content json")
	}

	toEdit, err := sqlboiler.FindEvent(ctx, bs.db, receivedEvent.ReferrerID.String)
	if err != nil {
		return result, merror.Internal().Describe("retrieving event to edit")
	}

	if receivedEvent.SenderID != toEdit.SenderID {
		return result, merror.Forbidden()
	}

	isDeleted, err := events.IsDeleted(toEdit)
	if err != nil {
		return result, merror.Transform(err).Describe("checking if event is deleted")
	}
	if isDeleted {
		return result, merror.Gone().Describe("cannot edit a deleted event")
	}

	if toEdit.Type != "msg.text" {
		return result, merror.Unauthorized().
			Describef("cannot edit events of type \"%s\" (only \"msg.text\")", toEdit.Type)
	}

	newContent := &events.MsgTextContent{
		Encrypted:    content.NewEncrypted,
		PublicKey:    content.NewPublicKey,
		LastEditedAt: null.TimeFrom(time.Now()),
	}

	newContentBytes, err := json.Marshal(newContent)
	if err != nil {
		return result, merror.Transform(err).Describe("marshalling new event content")
	}
	toEdit.Content = null.JSONFrom(newContentBytes)

	tx, err := bs.db.BeginTx(ctx, nil)
	if err != nil {
		return result, merror.Transform(err).Describe("unmarshaling content json")
	}

	rowsAff, err := toEdit.Update(ctx, tx, boil.Infer())
	if err != nil {
		atomic.SQLRollback(ctx, tx, err)
		return result, merror.Transform(err).Describe("updating event")
	}
	if rowsAff != 1 {
		atomic.SQLRollback(ctx, tx, err)
		return result, merror.Internal().Describef("%d rows affected", rowsAff)
	}

	err = toEdit.Reload(ctx, tx)
	if err != nil {
		atomic.SQLRollback(ctx, tx, err)
		return result, merror.Transform(err).Describe("reloading event")
	}

	err = tx.Commit()
	if err != nil {
		return result, merror.Transform(err).Describe("committing transaction")
	}

	event := events.FromSQLBoiler(toEdit)

	for _, after := range handler.After {
		if err := after(ctx, &receivedEvent, bs.db, bs.redConn, bs.identities); err != nil {
			// we log the error but we don’t return it
			logger.FromCtx(ctx).Warn().Err(err).Msgf("after %s event", receivedEvent.Type)
		}
	}

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
