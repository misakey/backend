package application

import (
	"context"
	"strconv"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type GetBackupCmd struct {
	AccountID string `json:"-"`
}

func (cmd GetBackupCmd) Validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.AccountID, v.Required, is.UUIDv4.Error("account id must be an uuid v4")),
	)
}

type GetBackupView struct {
	Data    string `json:"data"`
	Version int    `json:"version"`
}

func (sso SSOService) GetBackup(ctx context.Context, cmd GetBackupCmd) (GetBackupView, error) {
	view := GetBackupView{}

	// check access using context
	if err := sso.hasAccountAccess(ctx, cmd.AccountID); err != nil {
		return view, err
	}

	account, err := sso.accountService.Get(ctx, cmd.AccountID)
	if err != nil {
		return view, err
	}

	view.Data = account.BackupData
	view.Version = account.BackupVersion
	return view, nil
}

type UpdateBackupCmd struct {
	AccountID  string `json:"-"`
	Data       string `json:"data"`
	NewVersion int    `json:"version"`
}

func (cmd UpdateBackupCmd) Validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.AccountID, v.Required, is.UUIDv4.Error("account id must be an uuid v4")),
		v.Field(&cmd.Data, v.Required, is.Base64.Error("data must be base64 encoded")),
		v.Field(&cmd.NewVersion, v.Required),
	)
}

func (sso SSOService) UpdateBackup(ctx context.Context, cmd UpdateBackupCmd) error {
	// check access
	if err := sso.hasAccountAccess(ctx, cmd.AccountID); err != nil {
		return err
	}

	// retrieve the current state of the account
	currentAccount, err := sso.accountService.Get(ctx, cmd.AccountID)
	if err != nil {
		return err
	}

	expectedVersion := currentAccount.BackupVersion + 1
	if cmd.NewVersion != expectedVersion {
		return merror.Conflict().
			Describe("wrong new version value").
			Detail("version", merror.DVInvalid).
			Detail("expected_version", strconv.Itoa(expectedVersion))
	}

	currentAccount.BackupData = cmd.Data
	currentAccount.BackupVersion = cmd.NewVersion
	return sso.accountService.Update(ctx, &currentAccount)
}

// hasAccountAccess returns a forbidden error if the current context doesn't include
// an access claims with a subject identity linked to the given account ID
func (sso SSOService) hasAccountAccess(ctx context.Context, accountID string) error {
	// retrieve access claims
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}

	// the identity inside the accesses should be linked to the received account
	identity, err := sso.identityService.Get(ctx, acc.Subject)
	if err != nil {
		return err
	}

	// the identiy should not be orphan
	if !identity.AccountID.Valid {
		return merror.Forbidden()
	}

	// the identity linked account must match the received account id
	if identity.AccountID.String != accountID {
		return merror.Forbidden()
	}

	return nil
}
