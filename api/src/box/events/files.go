package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

// SetSavedStatus on file events contents for identity identityID
// the function alters events in fileEvents and only returns an error
func SetSavedStatus(ctx context.Context, exec boil.ContextExecutor, identityID string, fileEvents []*Event) error {
	// build ids array and indexed array
	var ids []string
	indexedEvents := make(map[string]*Event)
	for _, e := range fileEvents {
		var content MsgFileContent
		if err := content.Unmarshal(e.JSONContent); err != nil {
			return err
		}
		if content.EncryptedFileID != "" {
			ids = append(ids, content.EncryptedFileID)
			indexedEvents[content.EncryptedFileID] = e
		}
	}
	filters := files.SavedFileFilters{
		IdentityID:       identityID,
		EncryptedFileIDs: ids,
	}

	savedFiles, err := files.ListSavedFiles(ctx, exec, filters)
	if err != nil {
		return merr.From(err).Desc("getting saved files")
	}

	for _, file := range savedFiles {
		if e, ok := indexedEvents[file.EncryptedFileID]; ok {
			// set saved status to true
			var content MsgFileContent
			if err := content.Unmarshal(e.JSONContent); err != nil {
				return err
			}
			content.IsSaved = true
			if err := e.JSONContent.Marshal(content); err != nil {
				return merr.From(err).Descf("marshalling %s content", e.Type)
			}
		}
	}

	return nil
}

// IsFileOrphan ...
func IsFileOrphan(ctx context.Context, exec boil.ContextExecutor, fileID string) (bool, error) {
	// check that there is no saved file referring this file
	filters := files.SavedFileFilters{
		EncryptedFileIDs: []string{fileID},
	}

	savedFiles, err := files.ListSavedFiles(ctx, exec, filters)
	if err != nil {
		return false, err
	}
	if len(savedFiles) != 0 {
		return false, nil
	}

	// check that there is no box event referring this file
	filePartialEvents, err := list(ctx, exec, eventFilters{
		idOnly: true,
		eType:  null.StringFrom(etype.Msgfile),
		fileID: null.StringFrom(fileID),
	})
	if err != nil {
		return false, err
	}
	// if no event found, the file is orphan - should not happen
	if len(filePartialEvents) == 0 {
		return true, nil
	}

	// build list of ids to see if all of them have been deleted
	ids := make([]string, len(filePartialEvents))
	for i, e := range filePartialEvents {
		ids[i] = e.ID
	}
	deletePartialEvents, err := list(ctx, exec, eventFilters{
		idOnly:      true,
		eType:       null.StringFrom(etype.Msgdelete),
		referrerIDs: ids,
	})
	if err != nil {
		return false, err
	}
	// if at least one file events is not referred by a delete event, the file is not orphan
	if len(deletePartialEvents) != len(filePartialEvents) {
		return false, nil
	}

	// the file is orphan
	return true, nil
}

// HasAccessOrHasSavedFile ...
func HasAccessOrHasSavedFile(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client,
	identityID string, fileID string,
) (bool, error) {
	// 1. identity has access to files contained in boxes they have access to
	hasAccess, err := HasAccessToFile(ctx, exec, redConn, identityID, fileID)
	if err != nil {
		return false, err
	}
	if hasAccess {
		return true, nil
	}

	// 2. identity has access to files they have saved
	filters := files.SavedFileFilters{
		IdentityID:       identityID,
		EncryptedFileIDs: []string{fileID},
	}
	linkedSavedFiles, err := files.ListSavedFiles(ctx, exec, filters)
	if err != nil {
		return false, err
	}
	for _, savedFile := range linkedSavedFiles {
		if savedFile.IdentityID == identityID {
			return true, nil
		}
	}
	return false, nil
}

// HasAccessToFile if the file is in a box the identity have access to
func HasAccessToFile(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client,
	identityID string, fileID string,
) (bool, error) {
	// get all msg events mentioning the file
	filePartialEvents, err := list(ctx, exec, eventFilters{
		boxIDOnly: true,
		eType:     null.StringFrom(etype.Msgfile),
		fileID:    null.StringFrom(fileID),
	})
	if err != nil {
		return false, err
	}
	// for each file event, check the user has currently access to the box
	for _, event := range filePartialEvents {
		err := MustBeMember(ctx, exec, redConn, event.BoxID, identityID)
		// if no error has been raised, the access is allowed
		if err == nil {
			return true, nil
		}

		// if the error is not a forbidden, return it. Otherwise ignore it and continue checking
		if err != nil && !merr.IsAForbidden(err) {
			return false, err
		}
	}
	return false, nil
}

// DeleteOrphanFiles ...
func DeleteOrphanFiles(ctx context.Context, exec boil.ContextExecutor, filesRepo files.FileStorageRepo, fileIDs []string) error {
	for _, fileID := range fileIDs {
		// we need to check the existency of fileID
		// since it is set to "" when msg.delete is called on the msg.file
		if fileID != "" {
			isOrphan, err := IsFileOrphan(ctx, exec, fileID)
			if err != nil {
				return merr.From(err).Desc("checking file is orphan")
			}
			if isOrphan {
				if err := files.Delete(ctx, exec, filesRepo, fileID); err != nil {
					return merr.From(err).Desc("deleting stored file")
				}
			}
		}
	}

	return nil
}
