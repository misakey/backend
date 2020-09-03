package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
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

func (bs *BoxApplication) ListBoxes(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*ListBoxesRequest)
	// default limit is 10
	if req.Limit == 0 {
		req.Limit = 10
	}

	// retrieve accesses to filters boxes to return
	acc := ajwt.GetAccesses(ctx)
	boxes, err := boxes.ListJoinedAndAccessible(
		ctx,
		bs.db, bs.redConn, bs.identities,
		acc.IdentityID,
		req.Limit, req.Offset,
	)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting sender boxes")
	}

	return boxes, nil
}
