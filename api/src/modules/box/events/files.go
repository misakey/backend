package events

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
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
	boxEvents, err := FindByEncryptedFileID(ctx, exec, fileID)
	if err != nil && !merror.HasCode(err, merror.NotFoundCode) {
		return false, err
	}
	if len(boxEvents) != 0 {
		return false, nil
	}

	// the file is orphan
	return true, nil
}
