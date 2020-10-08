package application

import (
	"context"
	"strconv"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

type BackupQuery struct {
	accountID string
}

func (query *BackupQuery) BindAndValidate(eCtx echo.Context) error {
	query.accountID = eCtx.Param("id")
	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4),
	)
}

type BackupView struct {
	Data    string `json:"data"`
	Version int    `json:"version"`
}

// Handles GET /accounts/:id/backup - get the account backup information
func (sso *SSOService) GetBackup(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*BackupQuery)
	view := BackupView{}

	// check access using context
	acc := oidc.GetAccesses(ctx)
	if acc == nil ||
		acc.AccountID.IsZero() ||
		acc.AccountID.String != query.accountID {
		return view, merror.Forbidden()
	}

	account, err := sso.accountService.Get(ctx, query.accountID)
	if err != nil {
		return view, err
	}

	view.Data = account.BackupData
	view.Version = account.BackupVersion
	return view, nil
}

type BackupUpdateCmd struct {
	accountID  string
	Data       string `json:"data"`
	NewVersion int    `json:"version"`
}

func (cmd *BackupUpdateCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	cmd.accountID = eCtx.Param("id")

	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.accountID, v.Required, is.UUIDv4),
		v.Field(&cmd.Data, v.Required),
		v.Field(&cmd.NewVersion, v.Required),
	); err != nil {
		return merror.Transform(err).Describe("validating backup update cmd")
	}
	return nil
}

// Handles PUT /accounts/:id/backup - update the account backup information
func (sso *SSOService) UpdateBackup(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*BackupUpdateCmd)

	// check access using context
	acc := oidc.GetAccesses(ctx)
	if acc == nil ||
		acc.AccountID.IsZero() ||
		acc.AccountID.String != cmd.accountID {
		return nil, merror.Forbidden()
	}

	// retrieve the current state of the account
	currentAccount, err := sso.accountService.Get(ctx, cmd.accountID)
	if err != nil {
		return nil, err
	}

	expectedVersion := currentAccount.BackupVersion + 1
	if cmd.NewVersion != expectedVersion {
		return nil, merror.Conflict().
			Describe("wrong new version value").
			Detail("version", merror.DVInvalid).
			Detail("expected_version", strconv.Itoa(expectedVersion))
	}

	currentAccount.BackupData = cmd.Data
	currentAccount.BackupVersion = cmd.NewVersion
	return nil, sso.accountService.Update(ctx, &currentAccount)
}
