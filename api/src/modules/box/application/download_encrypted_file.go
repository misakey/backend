package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
	e "gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

type DownloadEncryptedFileRequest struct {
	fileID string
}

func (req *DownloadEncryptedFileRequest) BindAndValidate(eCtx echo.Context) error {
	req.fileID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.fileID, v.Required, is.UUIDv4),
	)
}

func (bs *BoxApplication) DownloadEncryptedFile(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*DownloadEncryptedFileRequest)

	// check the file does exist
	_, err := files.Get(ctx, bs.db, req.fileID)
	if err != nil {
		return nil, merror.Transform(err).Describe("finding msg.file event")
	}

	// check accesses
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}
	allowed, err := hasAccessToFile(ctx, bs.db, bs.redConn, bs.identities, req.fileID, acc.IdentityID)
	if err != nil {
		return nil, merror.Transform(err).Describe("checking access to file")
	}
	if !allowed {
		return nil, merror.Forbidden()
	}

	// download the file then render it
	data, err := files.Download(ctx, bs.filesRepo, req.fileID)
	if err != nil {
		return nil, merror.Transform(err).Describe("downloading")
	}

	return data, nil
}

func hasAccessToFile(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client, identities e.IdentityIntraprocessInterface,
	fileID string, identityID string,
) (bool, error) {
	// 1. identity has access to files contained in boxes they have access to
	// get all entities linked to the file
	// TODO (perf): list only boxes instead of all events
	linkedEvents, err := events.FindByEncryptedFileID(ctx, exec, fileID)
	// not finding an event with the encrypted file id doesn't mean the user doesn't have
	// an access to it - the file can be a saved file
	if err != nil && !merror.HasCode(err, merror.NotFoundCode) {
		return false, err
	}
	for _, event := range linkedEvents {
		err := events.MustMemberHaveAccess(ctx, exec, redConn, identities, event.BoxID, identityID)
		// if no error has been raised, the access is allowed
		if err == nil {
			return true, nil
		}

		// if the error is not a forbidden, return it otherwise ignore it and continue checking
		if !merror.HasCode(err, merror.ForbiddenCode) {
			return false, err
		}
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
