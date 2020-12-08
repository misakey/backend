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

// ListSavedFilesRequest ...
type ListSavedFilesRequest struct {
	Offset     *int   `query:"offset" json:"-"`
	Limit      *int   `query:"limit" json:"-"`
	IdentityID string `query:"identity_id" json:"-"`
}

// BindAndValidate ...
func (req *ListSavedFilesRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
		v.Field(&req.Offset, v.Min(0)),
		v.Field(&req.Limit, v.Min(0)),
	)
}

// ListSavedFiles ...
func (app *BoxApplication) ListSavedFiles(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListSavedFilesRequest)

	access := oidc.GetAccesses(ctx)
	if access == nil {
		return nil, merror.Unauthorized()
	}

	// check identity
	if req.IdentityID != access.IdentityID {
		return nil, merror.Forbidden().Detail("identity_id", merror.DVForbidden)
	}

	filters := files.SavedFileFilters{
		IdentityID: req.IdentityID,
		Offset:     req.Offset,
		Limit:      req.Limit,
	}

	return files.ListSavedFiles(ctx, app.DB, filters)
}
