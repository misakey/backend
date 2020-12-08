package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// AckNewEventsCountRequest ...
type AckNewEventsCountRequest struct {
	boxID string

	IdentityID string `json:"identity_id"`
}

// BindAndValidate ...
func (req *AckNewEventsCountRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
	)
}

// AckNewEventsCount ...
func (app *BoxApplication) AckNewEventsCount(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*AckNewEventsCountRequest)

	acc := oidc.GetAccesses(ctx)
	if acc.IdentityID != req.IdentityID {
		return nil, merror.Forbidden()
	}

	if err := events.DelCounts(ctx, app.RedConn, req.IdentityID, req.boxID); err != nil {
		return nil, err
	}

	return nil, nil
}
