package account

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type accountRepo interface {
	Create(context.Context, *domain.Account) error
	Get(context.Context, string) (*domain.Account, error)
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
	return acs.accounts.Create(ctx, account)
}

func (acs AccountService) Get(ctx context.Context, id string) (*domain.Account, error) {
	return acs.accounts.Get(ctx, id)
}
