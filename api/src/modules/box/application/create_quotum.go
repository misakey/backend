package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/quota"
)

type CreateQuotumRequest struct {
	IdentityID string `json:"identity_id"`
	Value      int64  `json:"value"`
	Origin     string `json:"origin"`
}

func (req *CreateQuotumRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
		v.Field(&req.Value, v.Required),
		v.Field(&req.Origin, v.Required),
	)
}

// only call be intraprocess entrypoints so no check required
func (app *BoxApplication) CreateQuotum(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CreateQuotumRequest)

	id, err := uuid.NewString()
	if err != nil {
		return nil, merror.Transform(err).Describe("generating uuid")
	}

	quotum := quota.Quotum{
		ID:         id,
		IdentityID: req.IdentityID,
		Value:      req.Value,
		Origin:     req.Origin,
	}

	if err := quota.Create(ctx, app.DB, &quotum); err != nil {
		return nil, merror.Transform(err).Describe("creating quotum")
	}

	return quotum, nil
}
