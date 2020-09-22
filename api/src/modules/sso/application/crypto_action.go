package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type ListCryptoActionsQuery struct {
	AccountID string
}

func (query ListCryptoActionsQuery) Validate() error {
	return v.ValidateStruct(&query,
		v.Field(&query.AccountID, v.Required, is.UUIDv4.Error("account id must be uuid v4")),
	)
}

type CryptoActionView struct {
	domain.CryptoAction
}

func (sso SSOService) ListCryptoActions(ctx context.Context, query ListCryptoActionsQuery) ([]CryptoActionView, error) {
	acc := ajwt.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return nil, merror.Forbidden()
	}

	if query.AccountID != acc.AccountID.String {
		return nil, merror.Forbidden().Describe("can only list one's own crypto actions")
	}

	actions, err := sso.cryptoActionService.ListCryptoActions(ctx, query.AccountID)
	if err != nil {
		return nil, err
	}

	views := make([]CryptoActionView, len(actions))
	for i, action := range actions {
		views[i].CryptoAction = action
	}

	return views, nil
}

type DeleteCryptoActionsQuery struct {
	AccountID     string
	UntilActionID string `json:"until_action_id"`
}

func (query DeleteCryptoActionsQuery) Validate() error {
	return v.ValidateStruct(&query,
		v.Field(&query.AccountID, v.Required, is.UUIDv4.Error("account id must be uuid v4")),
		v.Field(&query.UntilActionID, v.Required, is.UUIDv4.Error("until_action_id must be uuid v4")),
	)
}

func (sso SSOService) DeleteCryptoActionsUntil(ctx context.Context, query DeleteCryptoActionsQuery) error {
	acc := ajwt.GetAccesses(ctx)
	// querier must have an account
	if acc == nil || acc.AccountID.IsZero() {
		return merror.Forbidden()
	}

	if query.AccountID != acc.AccountID.String {
		return merror.Forbidden().Describe("can only delete one's own crypto actions")
	}

	action, err := sso.cryptoActionService.GetCryptoAction(ctx, query.UntilActionID)
	if err != nil {
		return merror.Transform(err).Describe("retrieving action")
	}

	if action.AccountID != query.AccountID {
		// We pretend not to have found the action
		return merror.NotFound().Describef("no action with ID %s", query.UntilActionID)
	}

	err = sso.cryptoActionService.DeleteCryptoActionsUntil(ctx, query.AccountID, action.CreatedAt)
	if err != nil {
		return merror.Transform(err).Describe("deleting actions")
	}

	return nil
}
