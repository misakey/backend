package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
)

type CountEventsRequest struct {
	boxID string
}

func (req *CountEventsRequest) BindAndValidate(eCtx echo.Context) error {
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
	)
}

func (app *BoxApplication) CountEvents(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CountEventsRequest)
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	if err := events.MustHaveAccess(ctx, app.DB, identityMapper, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	count, err := events.CountByBoxID(ctx, app.DB, req.boxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("counting boxes events")
	}

	return count, nil
}
