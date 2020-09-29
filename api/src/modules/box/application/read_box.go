package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

type ReadBoxRequest struct {
	boxID string
}

func (req *ReadBoxRequest) BindAndValidate(eCtx echo.Context) error {
	req.boxID = eCtx.Param("id")

	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
	)
}

func (bs *BoxApplication) ReadBox(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ReadBoxRequest)

	// check accesses
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}
	if err := events.MustMemberHaveAccess(ctx, bs.DB, bs.RedConn, bs.Identities, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	return boxes.Get(ctx, bs.DB, bs.Identities, req.boxID)
}
