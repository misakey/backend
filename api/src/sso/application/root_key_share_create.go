package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// RootKeyShareCreateCmd ...
type RootKeyShareCreateCmd struct {
	Share          string `json:"share"`
	OtherShareHash string `json:"other_share_hash"`
}

// BindAndValidate ...
func (cmd *RootKeyShareCreateCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
	}

	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.OtherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&cmd.Share, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	); err != nil {
		return merr.From(err).Desc("validating create root key share command")
	}
	return nil
}

// CreateRootKeyShare ...
func (sso *SSOService) CreateRootKeyShare(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*RootKeyShareCreateCmd)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}

	// the request must bear authorization for an account
	identity, err := identity.Get(ctx, sso.sqlDB, acc.IdentityID)
	if err != nil {
		return nil, err
	}
	if identity.AccountID.IsZero() {
		return nil, merr.Conflict().Desc("no account id in authorization").Add("account_id", merr.DVConflict)
	}

	rootKeyShare := crypto.RootKeyShare{
		AccountID:      identity.AccountID.String,
		Share:          cmd.Share,
		OtherShareHash: cmd.OtherShareHash,
	}
	err = crypto.CreateRootKeyShare(ctx, sso.redConn, rootKeyShare, sso.rootKeyShareExpirationTime)
	if err != nil {
		return nil, merr.From(err).Desc("creating key share")
	}
	return rootKeyShare, nil
}

// RootKeyShareQuery ...
type RootKeyShareQuery struct {
	otherShareHash string
}

// BindAndValidate ...
func (query *RootKeyShareQuery) BindAndValidate(eCtx echo.Context) error {
	query.otherShareHash = eCtx.Param("other-share-hash")

	if err := v.ValidateStruct(query,
		v.Field(&query.otherShareHash, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
	); err != nil {
		return merr.From(err).Desc("validating root key share query")
	}
	return nil
}

// GetRootKeyShare ...
func (sso *SSOService) GetRootKeyShare(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*RootKeyShareQuery)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}
	if acc.AccountID.IsZero() {
		return nil, merr.Conflict().
			Desc("no account id in authorization").
			Add("account_id", merr.DVConflict)
	}

	share, err := crypto.GetRootKeyShare(ctx, sso.redConn, query.otherShareHash)
	if err != nil {
		return nil, merr.From(err).Desc("getting key share")
	}

	if acc.AccountID.String != share.AccountID {
		return nil, merr.NotFound()
	}

	return share, nil
}
