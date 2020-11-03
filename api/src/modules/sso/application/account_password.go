package application

import (
	"context"
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn/argon2"
)

type PwdParamsQuery struct {
	accountID string
}

func (query *PwdParamsQuery) BindAndValidate(eCtx echo.Context) error {
	query.accountID = eCtx.Param("id")
	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4),
	)
}

type PwdParamsView struct {
	argon2.Params
}

func (sso *SSOService) GetAccountPwdParams(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*PwdParamsQuery)

	view := PwdParamsView{}

	account, err := sso.accountService.Get(ctx, query.accountID)
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
	accountID string

	OldPassword   argon2.HashedPassword `json:"old_prehashed_password"`
	NewPassword   argon2.HashedPassword `json:"new_prehashed_password"`
	BackupData    string                `json:"backup_data"`
	BackupVersion int                   `json:"backup_version"`
}

func (cmd *ChangePasswordCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	cmd.accountID = eCtx.Param("id")
	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.OldPassword),
		v.Field(&cmd.NewPassword),
		v.Field(&cmd.BackupData, v.Required),
		v.Field(&cmd.BackupVersion, v.Required),
		v.Field(&cmd.accountID, v.Required, is.UUIDv4),
	); err != nil {
		return merror.Transform(err).Describe("validating change password command")
	}
	return nil
}

func (sso *SSOService) ChangePassword(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*ChangePasswordCmd)

	// grab accesses from context
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Forbidden()
	}

	// verify authenticated account id is linked to the given account id
	if acc.AccountID.IsZero() || acc.AccountID.String != cmd.accountID {
		return nil, merror.Forbidden().Detail("account_id", merror.DVForbidden)
	}

	// get account
	account, err := sso.accountService.Get(ctx, cmd.accountID)
	if err != nil {
		return nil, err
	}

	// check old password
	oldPasswordValid, err := cmd.OldPassword.Matches(account.Password)
	if err != nil {
		return nil, err
	}
	if !oldPasswordValid {
		return nil, merror.Forbidden().Describe("invalid old password").Detail("old_password", merror.DVInvalid)
	}

	// update password
	account.Password, err = cmd.NewPassword.Hash()
	if err != nil {
		return nil, err
	}

	// check and update backup data
	if cmd.BackupVersion != account.BackupVersion+1 {
		return nil, merror.
			Conflict().
			Describe("bad backup version number").
			Detail("version", "invalid").
			Detail("expected_version", fmt.Sprintf("%d", account.BackupVersion+1))
	}

	account.BackupData = cmd.BackupData
	account.BackupVersion = cmd.BackupVersion

	// save account
	return nil, sso.accountService.Update(ctx, &account)
}

type PasswordResetCmd struct {
	Password   argon2.HashedPassword `json:"prehashed_password"`
	BackupData string                `json:"backup_data"`
}

func (cmd PasswordResetCmd) Validate() error {
	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.Password),
		v.Field(&cmd.BackupData, v.Required),
	); err != nil {
		return merror.Transform(err).Describe("validating reset password command")
	}
	return nil
}

func (sso *SSOService) resetPassword(ctx context.Context, cmd PasswordResetCmd, identityID string) error {
	// verify authenticated identity id is linked to the given account id
	identity, err := sso.identityService.Get(ctx, identityID)
	if err != nil {
		return merror.Forbidden().Describe("invalid token subject")
	}

	if identity.AccountID.String == "" {
		return merror.Conflict().
			Describe("identity is not linked to any account").
			Detail("identity_id", merror.DVConflict).
			Detail("account_id", merror.DVRequired)
	}

	// get account
	account, err := sso.accountService.Get(ctx, identity.AccountID.String)
	if err != nil {
		return err
	}

	backupArchive := domain.BackupArchive{
		AccountID: account.ID,
		Data:      null.StringFrom(account.BackupData),
	}
	err = sso.backupArchiveService.CreateBackupArchive(ctx, backupArchive)
	if err != nil {
		return err
	}

	// update password
	account.Password, err = cmd.Password.Hash()
	if err != nil {
		return err
	}

	account.BackupData = cmd.BackupData
	account.BackupVersion += 1

	// save account
	if err := sso.accountService.Update(ctx, &account); err != nil {
		return err
	}

	// create identity notification about password reset
	if err := sso.identityService.NotificationCreate(ctx, identity.ID, "user.reset_password", null.JSONFromPtr(nil)); err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("notifying identity %s", identity.ID)
	}
	return nil

}
