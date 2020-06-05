package account

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type accountRepo interface {
	Create(context.Context, *domain.Account) error
	Get(context.Context, string) (domain.Account, error)
	Update(context.Context, *domain.Account) error
}

type AccountService struct {
	accounts accountRepo
}

func NewAccountService(
	accountRepo accountRepo,
) AccountService {
	return AccountService{
		accounts: accountRepo,
	}
}

func (acs AccountService) Create(ctx context.Context, account *domain.Account) error {
	if err := acs.accounts.Create(ctx, account); err != nil {
		return merror.Transform(err).Describe("create account")
	}
	return nil
}

func (acs AccountService) Get(ctx context.Context, id string) (ret domain.Account, err error) {
	if ret, err = acs.accounts.Get(ctx, id); err != nil {
		return ret, merror.Transform(err).Describe("get account")
	}
	return ret, nil
}

func (acs AccountService) Update(ctx context.Context, account *domain.Account) error {
	if err := acs.accounts.Update(ctx, account); err != nil {
		return merror.Transform(err).Describe("update account")
	}
	return nil
}
