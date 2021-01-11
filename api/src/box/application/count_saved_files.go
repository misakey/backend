package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

// CountSavedFilesRequest ...
type CountSavedFilesRequest struct {
	IdentityID string `query:"identity_id" json:"-"`
}

// BindAndValidate ...
func (req *CountSavedFilesRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
	)
}

// CountSavedFiles ...
func (app *BoxApplication) CountSavedFiles(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CountSavedFilesRequest)

	access := oidc.GetAccesses(ctx)
	if access == nil {
		return nil, merr.Unauthorized()
	}

	// check identity
	if req.IdentityID != access.IdentityID {
		return nil, merr.Forbidden().Add("identity_id", merr.DVForbidden)
	}

	count, err := files.CountSavedFilesByIdentityID(ctx, app.DB, req.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("counting user saved files")
	}

	return count, nil
}
