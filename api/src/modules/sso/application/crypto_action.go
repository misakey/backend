package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

type ListCryptoActionsQuery struct {
	accountID string
}

func (query *ListCryptoActionsQuery) BindAndValidate(eCtx echo.Context) error {
	query.accountID = eCtx.Param("id")

	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4),
	)
}

type CryptoActionView struct {
	domain.CryptoAction
}

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

	actions, err := sso.cryptoActionService.ListCryptoActions(ctx, query.accountID)
	if err != nil {
		return nil, err
	}

	views := make([]CryptoActionView, len(actions))
	for i, action := range actions {
		views[i].CryptoAction = action
	}

	return views, nil
}

type GetCryptoActionQuery struct {
	accountID string
	actionID  string
}

func (query *GetCryptoActionQuery) BindAndValidate(eCtx echo.Context) error {
	query.accountID = eCtx.Param("account-id")
	query.actionID = eCtx.Param("action-id")

	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4),
		v.Field(&query.actionID, v.Required, is.UUIDv4),
	)
}

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

	action, err := sso.cryptoActionService.GetCryptoAction(ctx, query.actionID, query.accountID)
	if err != nil {
		return nil, err
	}

	view := CryptoActionView{
		CryptoAction: action,
	}

	return view, nil
}

type DeleteCryptoActionQuery struct {
	accountID string
	actionID  string
}

func (query *DeleteCryptoActionQuery) BindAndValidate(eCtx echo.Context) error {
	query.accountID = eCtx.Param("account-id")
	query.actionID = eCtx.Param("action-id")

	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4),
		v.Field(&query.actionID, v.Required, is.UUIDv4),
	)
}

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

	err := sso.cryptoActionService.DeleteCryptoAction(ctx, query.actionID, query.accountID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
