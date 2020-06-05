package application

import (
	"context"

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
