package application

import (
	"context"
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn/argon2"
)

type PwdParamsView struct {
	argon2.Params
}

func (sso SSOService) GetAccountPwdParams(ctx context.Context, query AccountQuery) (PwdParamsView, error) {
	view := PwdParamsView{}

	account, err := sso.accountService.Get(ctx, query.AccountID)
	if err != nil {
		return view, err
	}

	view.Params, err = argon2.DecodeParams(account.Password)
	if err != nil {
		return view, err
	}
	return view, nil
}

type ChangePasswordCmd struct {
	AccountID     string
	OldPassword   argon2.HashedPassword `json:"old_password"`
	NewPassword   argon2.HashedPassword `json:"new_password"`
	BackupData    string                `json:"backup_data"`
	BackupVersion int                   `json:"backup_version"`
}

func (cmd ChangePasswordCmd) Validate() error {
	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.OldPassword),
		v.Field(&cmd.NewPassword),
		v.Field(&cmd.BackupData, v.Required),
		v.Field(&cmd.BackupVersion, v.Required),
		v.Field(&cmd.AccountID, v.Required, is.UUIDv4.Error("account_id must be an uuid4")),
	); err != nil {
		return merror.Transform(err).Describe("validating change password command")
	}
	return nil
}

func (sso SSOService) ChangePassword(ctx context.Context, cmd ChangePasswordCmd) error {
	// grab accesses from context
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}

	// verify authenticated identity id is linked to the given account id
	identity, err := sso.identityService.Get(ctx, acc.Subject)
	if err != nil {
		return merror.Forbidden().Describe("invalid token subject")
	}

	if identity.AccountID.String != cmd.AccountID {
		return merror.Forbidden().Detail("account_id", merror.DVForbidden)
	}

	// get account
	account, err := sso.accountService.Get(ctx, cmd.AccountID)
	if err != nil {
		return err
	}

	// check old password
	oldPasswordValid, err := cmd.OldPassword.Matches(account.Password)
	if err != nil {
		return err
	}
	if !oldPasswordValid {
		return merror.Forbidden().Describe("invalid old password").Detail("old_password", merror.DVInvalid)
	}

	// update password
	account.Password, err = cmd.NewPassword.Hash()
	if err != nil {
		return err
	}

	// check and update backup data
	if cmd.BackupVersion != account.BackupVersion+1 {
		return merror.
			Conflict().
			Describe("bad backup version number").
			Detail("version", "invalid").
			Detail("expected_version", fmt.Sprintf("%d", account.BackupVersion+1))
	}

	account.BackupData = cmd.BackupData
	account.BackupVersion = cmd.BackupVersion

	// save account
	return sso.accountService.Update(ctx, &account)
}
