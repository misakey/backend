package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/realtime"
)

// UpdateBoxSettingsRequest ...
type UpdateBoxSettingsRequest struct {
	identityID string
	boxID      string

	Muted bool `json:"muted"`
}

// BindAndValidate ...
func (req *UpdateBoxSettingsRequest) BindAndValidate(eCtx echo.Context) error {
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

// UpdateBoxSettings ...
func (app *BoxApplication) UpdateBoxSettings(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*UpdateBoxSettingsRequest)

	acc := oidc.GetAccesses(ctx)
	if acc.IdentityID != req.identityID {
		return nil, merr.Forbidden()
	}

	// check box existency and membership
	if err := events.MustBeMember(ctx, app.DB, app.RedConn, req.boxID, req.identityID); err != nil {
		return nil, err
	}

	boxSetting := events.BoxSetting{
		IdentityID: req.identityID,
		BoxID:      req.boxID,
		Muted:      req.Muted,
	}

	createInfo, err := events.GetCreateInfo(ctx, app.DB, req.boxID)
	if err != nil {
		return nil, err
	}

	if err := events.UpdateBoxSetting(ctx, app.DB, boxSetting); err != nil {
		return nil, err
	}

	// remove the key used to send digests
	if req.Muted {
		if err := events.DelDigestCount(ctx, app.RedConn, req.identityID, req.boxID); err != nil {
			logger.FromCtx(ctx).Error().Err(err).Msg("could not delete digest key")
		}
	}

	// send the update to websockets
	realtime.SendUpdate(ctx, app.RedConn, acc.IdentityID, &realtime.Update{
		Type: "box.settings",
		Object: struct {
			events.BoxSetting
			OwnerOrgID string `json:"owner_org_id"`
		}{
			boxSetting, createInfo.OwnerOrgID,
		},
	})

	return boxSetting, nil
}
