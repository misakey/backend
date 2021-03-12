package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
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
		return merr.From(err).Ori(merr.OriBody)
	}
	return v.ValidateStruct(req,
		v.Field(&req.OtherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&req.Share, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&req.BoxID, v.Required, is.UUIDv4),
	)
}

// CreateKeyShare ...
func (app *BoxApplication) CreateKeyShare(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CreateKeyShareRequest)

	// check accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}
	if err := events.MustBeMember(ctx, app.DB, app.RedConn, req.BoxID, acc.IdentityID); err != nil {
		return nil, err
	}

	if err := keyshares.Create(
		ctx, app.DB,
		req.OtherShareHash, req.Share, req.EncryptedInvitationKeyShare.String, req.BoxID, acc.IdentityID,
	); err != nil {
		return nil, merr.From(err).Desc("creating key share")
	}

	return req, nil
}
