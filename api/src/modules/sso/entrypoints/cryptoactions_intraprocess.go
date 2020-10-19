package entrypoints

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/cryptoaction"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type CryptoActionIntraprocessInterface interface {
	Create(ctx context.Context, actions []domain.CryptoAction) error
}

type CryptoActionIntraprocess struct {
	service cryptoaction.CryptoActionService
}

func NewCryptoActionIntraprocess(service cryptoaction.CryptoActionService) CryptoActionIntraprocess {
	return CryptoActionIntraprocess{
		service: service,
	}
}

func (intraproc CryptoActionIntraprocess) Create(ctx context.Context, actions []domain.CryptoAction) error {
	return intraproc.service.CreateCryptoAction(ctx, actions)
}
