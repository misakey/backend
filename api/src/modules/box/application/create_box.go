package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

type CreateBoxRequest struct {
	PublicKey string `json:"public_key"`
	Title     string `json:"title"`
}

func (req *CreateBoxRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.PublicKey, v.Required),
		v.Field(&req.Title, v.Required, v.Length(1, 50)),
	)
}

func (bs *BoxApplication) CreateBox(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*CreateBoxRequest)

	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	// Check identity level
	if err := boxes.MustBeAtLeastLevel20(ctx, bs.db, bs.identities, acc.IdentityID); err != nil {
		return nil, merror.Transform(err).Describe("checking level")
	}

	event, err := events.CreateCreateEvent(ctx, req.Title, req.PublicKey, acc.IdentityID, bs.db)
	if err != nil {
		return nil, merror.Transform(err).Describe("creating create event")
	}

	// build the box view and return it
	box, err := boxes.Compute(ctx, event.BoxID, bs.db, bs.identities, &event)
	if err != nil {
		return nil, merror.Transform(err).Describe("building box")
	}
	return box, nil
}
