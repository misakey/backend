package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

type DeleteSavedFileRequest struct {
	ID string
}

func (req *DeleteSavedFileRequest) BindAndValidate(eCtx echo.Context) error {
	req.ID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.ID, v.Required, is.UUIDv4),
	)
}

func (bs *BoxApplication) DeleteSavedFile(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*DeleteSavedFileRequest)

	access := ajwt.GetAccesses(ctx)

	// get saved file
	savedFile, err := files.GetSavedFile(ctx, bs.db, req.ID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting saved file")
	}

	if savedFile.IdentityID != access.IdentityID {
		return nil, merror.Forbidden().Detail("id", merror.DVForbidden)
	}

	if err := files.DeleteSavedFile(ctx, bs.db, req.ID); err != nil {
		return nil, err
	}

	// delete stored file if orphan
	isOrphan, err := files.IsOrphan(ctx, bs.db, savedFile.EncryptedFileID)
	if err != nil {
		return nil, merror.Transform(err).Describe("deleting stored file")
	}
	if isOrphan {
		if err := files.Delete(ctx, bs.db, bs.filesRepo, savedFile.EncryptedFileID); err != nil {
			return nil, merror.Transform(err).Describe("deleting stored file")
		}
	}

	return nil, nil
}
