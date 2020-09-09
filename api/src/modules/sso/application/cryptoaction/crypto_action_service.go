package cryptoaction

import (
	"context"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type cryptoActionRepo interface {
	Get(ctx context.Context, actionID string) (domain.CryptoAction, error)
	List(ctx context.Context, accountID string) ([]domain.CryptoAction, error)
	Create(ctx context.Context, action domain.CryptoAction) error
	DeleteUntil(ctx context.Context, accountID string, untilTime time.Time) error
}

type CryptoActionService struct {
	cryptoActions cryptoActionRepo
}

func NewCryptoActionService(
	cryptoActionRepo cryptoActionRepo,
) CryptoActionService {
	return CryptoActionService{
		cryptoActions: cryptoActionRepo,
	}
}

func (service CryptoActionService) ListCryptoActions(ctx context.Context, accountID string) ([]domain.CryptoAction, error) {
	return service.cryptoActions.List(ctx, accountID)
}

func (service CryptoActionService) CreateCryptoAction(ctx context.Context, action domain.CryptoAction) error {
	return service.cryptoActions.Create(ctx, action)
}

func (service CryptoActionService) DeleteCryptoActionsUntil(ctx context.Context, accountID string, untilTime time.Time) error {
	return service.cryptoActions.DeleteUntil(ctx, accountID, untilTime)
}

func (service CryptoActionService) GetCryptoAction(ctx context.Context, actionID string) (domain.CryptoAction, error) {
	return service.cryptoActions.Get(ctx, actionID)
}
