package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// AuthBackupView ...
type AuthBackupView struct {
	BackupView

	AccountID string `json:"account_id"`
}

// GetBackupQuery ...
type GetBackupQuery struct {
	LoginChallenge string `query:"login_challenge"`
	IdentityID     string `query:"identity_id"`
}

// BindAndValidate ...
func (query *GetBackupQuery) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(query); err != nil {
		return merr.BadRequest().Ori(merr.OriQuery).Desc(err.Error())
	}

	return v.ValidateStruct(query,
		v.Field(&query.LoginChallenge, v.Required),
		v.Field(&query.IdentityID, v.Required, is.UUIDv4),
	)
}

// GetBackupDuringAuth ...
func (sso *SSOService) GetBackupDuringAuth(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*GetBackupQuery)
	view := AuthBackupView{}

	// get token
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return view, merr.Forbidden().Ori(merr.OriHeaders).Desc("bearer token could not be found")
	}

	// check login challenge
	if acc.Subject != query.LoginChallenge {
		return view, merr.Forbidden().Add("login_challenge", merr.DVInvalid)
	}

	// identity_id must match
	if acc.IdentityID != query.IdentityID {
		return view, merr.Forbidden().Add("identity_id", merr.DVForbidden)
	}

	// get identity
	curIdentity, err := identity.Get(ctx, sso.ssoDB, query.IdentityID)
	if err != nil {
		return view, err
	}

	if curIdentity.AccountID.IsZero() {
		return view, merr.Conflict().Desc("identity has no account")
	}

	account, err := identity.GetAccount(ctx, sso.ssoDB, curIdentity.AccountID.String)
	if err != nil {
		return view, err
	}

	view.Data = account.BackupData
	view.Version = account.BackupVersion
	view.AccountID = account.ID

	return view, nil
}
