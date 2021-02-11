package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	//	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

// SecretStorageView ...
type SecretStorageView struct {
	Secrets crypto.Secrets `json:"secrets"`

	AccountID string `json:"account_id"`
}

// GetSecretStorageQuery ...
type GetSecretStorageQuery struct {
	LoginChallenge string `query:"login_challenge"`
	IdentityID     string `query:"identity_id"`
}

// BindAndValidate ...
func (query *GetSecretStorageQuery) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(query); err != nil {
		return merr.BadRequest().Ori(merr.OriQuery).Desc(err.Error())
	}

	return v.ValidateStruct(query,
		v.Field(&query.LoginChallenge, v.Required),
		v.Field(&query.IdentityID, v.Required, is.UUIDv4),
	)
}

// GetSecretStorageDuringAuth ...
func (sso *SSOService) GetSecretStorageDuringAuth(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*GetSecretStorageQuery)

	// get token
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden().Ori(merr.OriHeaders).Desc("bearer token could not be found")
	}

	// check login challenge
	if acc.Subject != query.LoginChallenge {
		return nil, merr.Forbidden().Add("login_challenge", merr.DVInvalid)
	}

	// identity_id must match
	if acc.IdentityID != query.IdentityID {
		return nil, merr.Forbidden().Add("identity_id", merr.DVForbidden)
	}

	// get identity
	curIdentity, err := identity.Get(ctx, sso.sqlDB, query.IdentityID)
	if err != nil {
		return nil, err
	}

	if curIdentity.AccountID.IsZero() {
		return nil, merr.Conflict().Desc("identity has no account")
	}

	secrets, err := crypto.GetAccountSecrets(ctx, sso.sqlDB, curIdentity.AccountID.String)
	if err != nil {
		if err == crypto.ErrNoRootKey {
			return nil, merr.Conflict().Desc("Account has no root key; it requires migration.")
		}

		return nil, merr.From(err)
	}

	result := SecretStorageView{
		Secrets:   secrets,
		AccountID: curIdentity.AccountID.String,
	}

	return result, nil
}
