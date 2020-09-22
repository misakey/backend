package application

import (
	"context"
	"strconv"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type AccountQuery struct {
	AccountID string
}

func (query AccountQuery) Validate() error {
	return v.ValidateStruct(&query,
		v.Field(&query.AccountID, v.Required, is.UUIDv4.Error("account id must be an uuid v4")),
	)
}

type BackupView struct {
	Data    string `json:"data"`
	Version int    `json:"version"`
}

func (sso SSOService) GetBackup(ctx context.Context, query AccountQuery) (BackupView, error) {
	view := BackupView{}

	// check access using context
	acc := ajwt.GetAccesses(ctx)
	if acc == nil ||
		acc.AccountID.IsZero() ||
		acc.AccountID.String != query.AccountID {
		return view, merror.Forbidden()
	}

	account, err := sso.accountService.Get(ctx, query.AccountID)
	if err != nil {
		return view, err
	}

	view.Data = account.BackupData
	view.Version = account.BackupVersion
	return view, nil
}

type UpdateBackupCmd struct {
	accountID  string
	Data       string `json:"data"`
	NewVersion int    `json:"version"`
}

func (cmd *UpdateBackupCmd) SetAccountID(id string) {
	cmd.accountID = id
}

func (cmd UpdateBackupCmd) Validate() error {
	return v.ValidateStruct(&cmd,
		v.Field(&cmd.accountID, v.Required, is.UUIDv4.Error("account id must be an uuid v4")),
		v.Field(&cmd.Data, v.Required),
		v.Field(&cmd.NewVersion, v.Required),
	)
}

func (sso SSOService) UpdateBackup(ctx context.Context, cmd UpdateBackupCmd) error {
	// check access using context
	acc := ajwt.GetAccesses(ctx)
	if acc == nil ||
		acc.AccountID.IsZero() ||
		acc.AccountID.String != cmd.accountID {
		return merror.Forbidden()
	}

	// retrieve the current state of the account
	currentAccount, err := sso.accountService.Get(ctx, cmd.accountID)
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
