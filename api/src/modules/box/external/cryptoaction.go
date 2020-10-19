package external

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type CryptoActionRepo interface {
	Create(ctx context.Context, actions []domain.CryptoAction) error
}
