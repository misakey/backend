package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// ListCryptoActionsQuery ...
type ListCryptoActionsQuery struct {
	accountID string
}

// BindAndValidate ...
func (query *ListCryptoActionsQuery) BindAndValidate(eCtx echo.Context) error {
	query.accountID = eCtx.Param("id")

	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4),
	)
}

// CryptoActionView ...
type CryptoActionView struct {
	crypto.Action
}

// ListCryptoActions ...
func (sso *SSOService) ListCryptoActions(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*ListCryptoActionsQuery)

	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merror.Forbidden()
	}

	if query.accountID != acc.AccountID.String {
		return nil, merror.Forbidden().Describe("can only list one's own crypto actions")
	}

	actions, err := crypto.ListActions(ctx, sso.sqlDB, query.accountID)
	if err != nil {
		return nil, err
	}

	views := make([]CryptoActionView, len(actions))
	for i, action := range actions {
		views[i].Action = action
	}

	return views, nil
}

// GetCryptoActionQuery ...
type GetCryptoActionQuery struct {
	accountID string
	actionID  string
}

// BindAndValidate ...
func (query *GetCryptoActionQuery) BindAndValidate(eCtx echo.Context) error {
	query.accountID = eCtx.Param("account-id")
	query.actionID = eCtx.Param("action-id")

	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4),
		v.Field(&query.actionID, v.Required, is.UUIDv4),
	)
}

// GetCryptoAction ...
func (sso *SSOService) GetCryptoAction(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*GetCryptoActionQuery)

	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merror.Forbidden()
	}

	if query.accountID != acc.AccountID.String {
		return nil, merror.Forbidden().Describe("can only get one's own crypto actions")
	}

	action, err := crypto.GetAction(ctx, sso.sqlDB, query.actionID, query.accountID)
	if err != nil {
		return nil, err
	}

	view := CryptoActionView{
		Action: action,
	}

	return view, nil
}

// DeleteCryptoActionQuery ...
type DeleteCryptoActionQuery struct {
	accountID string
	actionID  string
}

// BindAndValidate ...
func (query *DeleteCryptoActionQuery) BindAndValidate(eCtx echo.Context) error {
	query.accountID = eCtx.Param("account-id")
	query.actionID = eCtx.Param("action-id")

	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4),
		v.Field(&query.actionID, v.Required, is.UUIDv4),
	)
}

// DeleteCryptoAction ...
func (sso *SSOService) DeleteCryptoAction(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*DeleteCryptoActionQuery)

	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merror.Forbidden()
	}

	if query.accountID != acc.AccountID.String {
		return nil, merror.Forbidden().Describe("can only delete one's own crypto actions")
	}

	// start transaction since write actions will be performed
	tr, err := sso.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	err = crypto.DeleteAction(ctx, tr, query.actionID, query.accountID)
	if err != nil {
		return nil, err
	}
	return nil, tr.Commit()
}
