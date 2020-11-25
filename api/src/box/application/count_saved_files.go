package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

type CountSavedFilesRequest struct {
	IdentityID string `query:"identity_id" json:"-"`
}

func (req *CountSavedFilesRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
	)
}

func (app *BoxApplication) CountSavedFiles(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CountSavedFilesRequest)
	
	access := oidc.GetAccesses(ctx)
	if access == nil {
		return nil, merror.Unauthorized()
	}

	// check identity
	if req.IdentityID != access.IdentityID {
		return nil, merror.Forbidden().Detail("identity_id", merror.DVForbidden)
	}


	count, err := files.CountSavedFilesByIdentityID(ctx, app.DB, req.IdentityID)
	if err != nil {
		return nil, merror.Transform(err).Describe("counting user saved files")
	}

	return count, nil
}
