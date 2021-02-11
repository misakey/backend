package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn/argon2"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// PwdParamsQuery ...
type PwdParamsQuery struct {
	accountID string
}

// BindAndValidate ...
func (query *PwdParamsQuery) BindAndValidate(eCtx echo.Context) error {
	query.accountID = eCtx.Param("id")
	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4),
	)
}

// PwdParamsView ...
type PwdParamsView struct {
	argon2.Params
}

// GetAccountPwdParams ...
func (sso *SSOService) GetAccountPwdParams(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*PwdParamsQuery)

	view := PwdParamsView{}

	account, err := identity.GetAccount(ctx, sso.sqlDB, query.accountID)
	if err != nil {
		return view, err
	}

	view.Params, err = argon2.DecodeParams(account.Password)
	if err != nil {
		return view, err
	}
	return view, nil
}

// ChangePasswordCmd ...
type ChangePasswordCmd struct {
	accountID string

	OldPassword             argon2.HashedPassword `json:"old_prehashed_password"`
	NewPassword             argon2.HashedPassword `json:"new_prehashed_password"`
	EncryptedAccountRootKey string                `json:"encrypted_account_root_key"`
}

// BindAndValidate ...
func (cmd *ChangePasswordCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
	}

	cmd.accountID = eCtx.Param("id")
	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.OldPassword),
		v.Field(&cmd.NewPassword),
		v.Field(&cmd.EncryptedAccountRootKey, v.Required, v.Match(format.UnpaddedURLSafeBase64)),
		v.Field(&cmd.accountID, v.Required, is.UUIDv4),
	); err != nil {
		return merr.From(err).Desc("validating change password command")
	}

	return nil
}

// ChangePassword ...
func (sso *SSOService) ChangePassword(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*ChangePasswordCmd)

	// grab accesses from context
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}

	// verify authenticated account id is linked to the given account id
	if acc.AccountID.IsZero() || acc.AccountID.String != cmd.accountID {
		return nil, merr.Forbidden().Add("account_id", merr.DVForbidden)
	}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	// get account
	account, err := identity.GetAccount(ctx, tr, cmd.accountID)
	if err != nil {
		return nil, err
	}

	// check old password
	oldPasswordValid, err := cmd.OldPassword.Matches(account.Password)
	if err != nil {
		return nil, err
	}
	if !oldPasswordValid {
		err = merr.Forbidden().Desc("invalid old password").Add("old_password", merr.DVInvalid)
		return nil, err
	}

	// update password
	account.Password, err = cmd.NewPassword.Hash()
	if err != nil {
		return nil, err
	}

	err = crypto.UpdateRootKey(ctx, tr, cmd.accountID, cmd.EncryptedAccountRootKey)
	if err != nil {
		return nil, err
	}

	// save account
	err = identity.UpdateAccount(ctx, tr, &account)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}
