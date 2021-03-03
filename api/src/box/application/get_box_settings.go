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

// GetBoxSettingsRequest ...
type GetBoxSettingsRequest struct {
	identityID string
	boxID      string
}

// BindAndValidate ...
func (req *GetBoxSettingsRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	req.identityID = eCtx.Param("id")
	req.boxID = eCtx.Param("bid")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.identityID, v.Required, is.UUIDv4),
	)
}

// GetBoxSettings ...
func (app *BoxApplication) GetBoxSettings(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*GetBoxSettingsRequest)

	acc := oidc.GetAccesses(ctx)
	if acc == nil || acc.IdentityID != req.identityID {
		return nil, merr.Forbidden()
	}

	// check box existency and access
	if err := events.MustBeMember(ctx, app.DB, app.RedConn, req.boxID, req.identityID); err != nil {
		return nil, err
	}

	return events.GetBoxSetting(ctx, app.DB, req.identityID, req.boxID)

}
