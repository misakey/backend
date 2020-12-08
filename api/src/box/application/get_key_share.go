package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/keyshares"
)

// GetKeyShareRequest ...
type GetKeyShareRequest struct {
	otherShareHash string
}

// BindAndValidate ...
func (req *GetKeyShareRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriPath)
	}
	req.otherShareHash = eCtx.Param("other-share-hash")
	return v.ValidateStruct(req,
		v.Field(&req.otherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	)
}

// GetKeyShare ...
func (app *BoxApplication) GetKeyShare(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*GetKeyShareRequest)

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// check accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	ks, err := keyshares.Get(ctx, app.DB, req.otherShareHash)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting key share")
	}

	if err := events.MustHaveAccess(ctx, app.DB, identityMapper, ks.BoxID, acc.IdentityID); err != nil {
		return nil, err
	}

	return ks, nil
}
