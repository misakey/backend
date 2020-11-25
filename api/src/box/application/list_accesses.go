package application

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

type ListAccessesRequest struct {
	boxID string
}

func (req *ListAccessesRequest) BindAndValidate(eCtx echo.Context) error {
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
	)
}

func (app *BoxApplication) ListAccesses(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListAccessesRequest)

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// retrieve accesses to filters boxes to return
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}
	if err := events.MustBeAdmin(ctx, app.DB, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	accessEvents, err := events.FindActiveAccesses(ctx, app.DB, req.boxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting sender accesses")
	}

	views := make([]events.View, len(accessEvents))
	for i, e := range accessEvents {
		// the user is admin and we need to have transparent identity to list them
		views[i], err = e.Format(ctx, identityMapper, true)
		if err != nil {
			return views, merror.Transform(err).Describe("computing access view")
		}
	}
	return views, nil
}
