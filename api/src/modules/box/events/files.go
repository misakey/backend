package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

func IsFileOrphan(ctx context.Context, exec boil.ContextExecutor, fileID string) (bool, error) {
	// check that there is no saved file referring this file
	savedFiles, err := files.ListSavedFilesByFileID(ctx, exec, fileID)
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

func HasAccessOrHasSavedFile(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper,
	identityID string, fileID string,
) (bool, error) {
	// 1. identity has access to files contained in boxes they have access to
	hasAccess, err := HasAccessToFile(ctx, exec, redConn, identities, identityID, fileID)
	if err != nil {
		return false, err
	}
	if hasAccess {
		return true, nil
	}

	// 2. identity has access to files they have saved
	// TODO (perf): filter directly by IdentityID = use a new function files.ListSaved(savedFileFilters{})
	linkedSavedFiles, err := files.ListSavedFilesByFileID(ctx, exec, fileID)
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

// identity has access to files contained in boxes they have access to
func HasAccessToFile(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper,
	identityID string, fileID string,
) (bool, error) {
	// get all msg events mentionning the file
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
		err := MustMemberHaveAccess(ctx, exec, redConn, identities, event.BoxID, identityID)
		// if no error has been raised, the access is allowed
		if err == nil {
			return true, nil
		}

		// if the error is not a forbidden, return it. Otherwise ignore it and continue checking
		if !merror.HasCode(err, merror.ForbiddenCode) {
			return false, err
		}
	}
	return false, nil
}
