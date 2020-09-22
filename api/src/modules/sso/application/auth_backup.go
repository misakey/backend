package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	//	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/oidc"
)

type AuthBackupView struct {
	BackupView

	AccountID string `json:"account_id"`
}

type GetBackupQuery struct {
	LoginChallenge string
	IdentityID     string
}

func (cmd GetBackupQuery) Validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.LoginChallenge, v.Required),
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4),
	)
}

func (sso SSOService) GetBackupDuringAuth(ctx context.Context, query GetBackupQuery) (AuthBackupView, error) {
	view := AuthBackupView{}

	// get token
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return view, merror.Forbidden().From(merror.OriHeaders).Describe("bearer token could not be found")
	}

	// check login challenge
	if acc.Subject != query.LoginChallenge {
		return view, merror.Forbidden().Detail("login_challenge", merror.DVInvalid)
	}

	// identity_id must match
	if acc.IdentityID != query.IdentityID {
		return view, merror.Forbidden().Detail("identity_id", merror.DVForbidden)
	}

	// process must be at least on ACR 2
	if err := acc.ACRIsGTE(ajwt.ACRSecLvl2); err != nil {
		return view, merror.Forbidden().Describe("acr must be at least 2")
	}

	// get identity
	identity, err := sso.identityService.Get(ctx, query.IdentityID)
	if err != nil {
		return view, err
	}

	if identity.AccountID.IsZero() {
		return view, merror.Conflict().Describe("identity has no account")
	}

	account, err := sso.accountService.Get(ctx, identity.AccountID.String)
	if err != nil {
		return view, err
	}

	view.Data = account.BackupData
	view.Version = account.BackupVersion
	view.AccountID = account.ID

	return view, nil
}
