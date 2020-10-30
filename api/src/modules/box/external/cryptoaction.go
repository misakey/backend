package external

import (
	"context"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type CryptoActionRepo interface {
	Create(ctx context.Context, actions []domain.CryptoAction) error
	CreateInvitationActions(ctx context.Context, senderID string, boxID string, identifierValue string, actionsData null.JSON) error
}
