package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
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

func (bs *BoxApplication) CountEvents(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*CountEventsRequest)
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	if err := events.MustHaveAccess(ctx, bs.db, bs.identities, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	count, err := events.CountByBoxID(ctx, bs.db, req.boxID)
	if err != nil {
		return nil, merror.Transform(err).Describe("counting sender boxes")
	}

	return count, nil
}
