package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
)

// CountBoxFilesRequest ...
type CountBoxFilesRequest struct {
	boxID string
}

// BindAndValidate ...
func (req *CountBoxFilesRequest) BindAndValidate(eCtx echo.Context) error {
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
	)
}

// CountBoxFiles ...
func (app *BoxApplication) CountBoxFiles(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CountBoxFilesRequest)
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}

	if err := events.MustBeMember(ctx, app.DB, app.RedConn, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	count, err := events.CountFilesByBoxID(ctx, app.DB, req.boxID)
	if err != nil {
		return nil, merr.From(err).Desc("counting boxes files")
	}

	return count, nil
}
