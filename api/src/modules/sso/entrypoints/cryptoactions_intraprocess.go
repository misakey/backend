package entrypoints

import (
	"context"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/cryptoaction"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type CryptoActionIntraprocessInterface interface {
	Create(ctx context.Context, actions []domain.CryptoAction) error
	CreateInvitationActions(ctx context.Context, senderID string, boxID string, identifierValue string, actionsData null.JSON) error
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

func (intraproc CryptoActionIntraprocess) CreateInvitationActions(ctx context.Context, senderID string, boxID string, identifierValue string, actionsData null.JSON) error {
	return intraproc.service.CreateInvitationActions(ctx, senderID, boxID, identifierValue, actionsData)
}
