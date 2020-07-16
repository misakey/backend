package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
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

func (bs *BoxApplication) ReadBox(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*ReadBoxRequest)

	box, err := boxes.Get(ctx, bs.db, bs.identities, req.boxID)
	if err != nil {
		return nil, err
	}

	return box, nil
}
