package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

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

func (app *BoxApplication) CreateBox(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CreateBoxRequest)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	event, err := events.CreateCreateEvent(
		ctx,
		req.Title,
		req.PublicKey,
		acc.IdentityID,
		app.DB,
		app.RedConn,
		identityMapper,
		app.filesRepo,
	)
	if err != nil {
		return nil, merror.Transform(err).Describe("creating create event")
	}

	// build the box view and return it
	box, err := events.Compute(ctx, event.BoxID, app.DB, identityMapper, &event)
	if err != nil {
		return nil, merror.Transform(err).Describe("building box")
	}

	return box, nil
}
