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

type ListSavedFilesRequest struct {
	IdentityID string `query:"identity_id" json:"-"`
}

func (req *ListSavedFilesRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
	)
}

func (bs *BoxApplication) ListSavedFiles(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*ListSavedFilesRequest)

	access := ajwt.GetAccesses(ctx)
	if access == nil {
		return nil, merror.Unauthorized()
	}

	// check identity
	if req.IdentityID != access.IdentityID {
		return nil, merror.Forbidden().Detail("identity_id", merror.DVForbidden)
	}

	return files.ListSavedFilesByIdentityID(ctx, bs.db, req.IdentityID)
}
