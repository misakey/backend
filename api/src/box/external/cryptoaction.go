package external

import (
	"context"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
)

type CryptoActionRepo interface {
	CreateCryptoActions(ctx context.Context, actions []crypto.Action) error
	CreateInvitationActions(ctx context.Context, senderID string, boxID string, boxTitle string, identifierValue string, actionsData null.JSON) error
}
