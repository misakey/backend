package external

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
)

// CryptoRepo ...
type CryptoRepo interface {
	CreateActions(ctx context.Context, actions []crypto.Action) error
}
