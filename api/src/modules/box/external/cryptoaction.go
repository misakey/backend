package external

import (
	"context"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type CryptoActionRepo interface {
	CreateCryptoAction(ctx context.Context, actions []domain.CryptoAction) error
	CreateInvitationActions(ctx context.Context, senderID string, boxID string, boxTitle string, identifierValue string, actionsData null.JSON) error
}
