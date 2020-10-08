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

type DeleteCryptoActionsCmd struct {
	accountID     string
	UntilActionID string `json:"until_action_id"`
}

func (query *DeleteCryptoActionsCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(query); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}
	query.accountID = eCtx.Param("id")

	return v.ValidateStruct(query,
		v.Field(&query.accountID, v.Required, is.UUIDv4.Error("account id must be uuid v4")),
		v.Field(&query.UntilActionID, v.Required, is.UUIDv4.Error("until_action_id must be uuid v4")),
	)
}

func (sso *SSOService) DeleteCryptoActionsUntil(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*DeleteCryptoActionsCmd)

	acc := oidc.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merror.Forbidden()
	}

	if cmd.accountID != acc.AccountID.String {
		return nil, merror.Forbidden().Describe("can only delete one's own crypto actions")
	}

	action, err := sso.cryptoActionService.GetCryptoAction(ctx, cmd.UntilActionID)
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving action")
	}

	if action.AccountID != cmd.accountID {
		// We pretend not to have found the action
		return nil, merror.NotFound().Describef("no action with ID %s", cmd.UntilActionID)
	}

	err = sso.cryptoActionService.DeleteCryptoActionsUntil(ctx, cmd.accountID, action.CreatedAt)
	if err != nil {
		return nil, merror.Transform(err).Describe("deleting actions")
	}

	return nil, nil
}
