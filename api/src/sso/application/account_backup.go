package application

import (
	"context"
	"strconv"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// BackupQuery ...
type BackupQuery struct {
	accountID string
}

// BindAndValidate ...
func (query *BackupQuery) BindAndValidate(eCtx echo.Context) error {
	query.accountID = eCtx.Param("id")
	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4),
	)
}

// BackupView ...
type BackupView struct {
	Data    string `json:"data"`
	Version int    `json:"version"`
}

// GetBackup handles GET /accounts/:id/backup
// Get the account backup information
func (sso *SSOService) GetBackup(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*BackupQuery)
	view := BackupView{}

	// check access using context
	acc := oidc.GetAccesses(ctx)
	if acc == nil ||
		acc.AccountID.IsZero() ||
		acc.AccountID.String != query.accountID {
		return view, merr.Forbidden()
	}

	account, err := identity.GetAccount(ctx, sso.ssoDB, query.accountID)
	if err != nil {
		return view, err
	}

	view.Data = account.BackupData
	view.Version = account.BackupVersion
	return view, nil
}

// BackupUpdateCmd ...
type BackupUpdateCmd struct {
	accountID  string
	Data       string `json:"data"`
	NewVersion int    `json:"version"`
}

// BindAndValidate ...
func (cmd *BackupUpdateCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.BadRequest().Ori(merr.OriBody).Desc(err.Error())
	}

	cmd.accountID = eCtx.Param("id")

	if err := v.ValidateStruct(cmd,
		v.Field(&cmd.accountID, v.Required, is.UUIDv4),
		v.Field(&cmd.Data, v.Required),
		v.Field(&cmd.NewVersion, v.Required),
	); err != nil {
		return merr.From(err).Desc("validating backup update cmd")
	}
	return nil
}

// UpdateBackup handles PUT /accounts/:id/backup
// Update the account backup information
func (sso *SSOService) UpdateBackup(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*BackupUpdateCmd)

	// check access using context
	acc := oidc.GetAccesses(ctx)
	if acc == nil ||
		acc.AccountID.IsZero() ||
		acc.AccountID.String != cmd.accountID {
		return nil, merr.Forbidden()
	}

	// start transaction since write actions will be performed
	tr, err := sso.ssoDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	// retrieve the current state of the account
	account, err := identity.GetAccount(ctx, tr, cmd.accountID)
	if err != nil {
		return nil, err
	}

	expectedVersion := account.BackupVersion + 1
	if cmd.NewVersion != expectedVersion {
		err = merr.Conflict().
			Desc("wrong new version value").
			Add("version", merr.DVInvalid).
			Add("expected_version", strconv.Itoa(expectedVersion))
		return nil, err
	}

	account.BackupData = cmd.Data
	account.BackupVersion = cmd.NewVersion
	err = identity.UpdateAccount(ctx, tr, &account)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}
