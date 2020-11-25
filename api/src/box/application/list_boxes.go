package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/boxes"
)

type ListBoxesRequest struct {
	Offset int `query:"offset" json:"-"`
	Limit  int `query:"limit" json:"-"`
}

func (req *ListBoxesRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriQuery)
	}
	return v.ValidateStruct(req,
		v.Field(&req.Offset, v.Min(0)),
		v.Field(&req.Limit, v.Min(0)),
	)
}

func (app *BoxApplication) ListBoxes(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListBoxesRequest)
	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// default limit is 10
	if req.Limit == 0 {
		req.Limit = 10
	}

	// retrieve accesses to filters boxes to return
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}
	boxes, err := boxes.ListSenderBoxes(
		ctx,
		app.DB, app.RedConn, identityMapper,
		acc.IdentityID,
		req.Limit, req.Offset,
	)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting sender boxes")
	}

	return boxes, nil
}
