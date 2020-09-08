package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/keyshares"
)

type GetKeyShareRequest struct {
	otherShareHash string
}

func (req *GetKeyShareRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriPath)
	}
	req.otherShareHash = eCtx.Param("other-share-hash")
	return v.ValidateStruct(req,
		v.Field(&req.otherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	)
}

func (bs *BoxApplication) GetKeyShare(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*GetKeyShareRequest)

	// check accesses
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	ks, err := keyshares.Get(ctx, bs.db, req.otherShareHash)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting key share")
	}

	if err := events.MustHaveAccess(ctx, bs.db, bs.identities, ks.BoxID, acc.IdentityID); err != nil {
		return nil, err
	}

	return ks, nil
}
