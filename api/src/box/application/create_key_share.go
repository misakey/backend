package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/keyshares"
)

// CreateKeyShareRequest ...
type CreateKeyShareRequest struct {
	keyshares.BoxKeyShare
}

// BindAndValidate ...
func (req *CreateKeyShareRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.OtherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&req.Share, v.Required, is.Base64),
		v.Field(&req.BoxID, v.Required, is.UUIDv4),
	)
}

// CreateKeyShare ...
func (app *BoxApplication) CreateKeyShare(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CreateKeyShareRequest)

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	// check accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}
	if err := events.MustMemberHaveAccess(ctx, app.DB, app.RedConn, identityMapper, req.BoxID, acc.IdentityID); err != nil {
		return nil, err
	}

	if err := keyshares.Create(
		ctx, app.DB,
		req.OtherShareHash, req.Share, req.EncryptedInvitationKeyShare.String, req.BoxID, acc.IdentityID,
	); err != nil {
		return nil, merror.Transform(err).Describe("creating key share")
	}

	return req, nil
}
