package cryptoaction

import (
	"context"
	"time"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type cryptoActionRepo interface {
	Get(ctx context.Context, actionID string) (domain.CryptoAction, error)
	List(ctx context.Context, accountID string) ([]domain.CryptoAction, error)
	Create(ctx context.Context, actions []domain.CryptoAction) error
	DeleteUntil(ctx context.Context, accountID string, untilTime time.Time) error
}

type CryptoActionService struct {
	cryptoActions   cryptoActionRepo
	identityService identity.IdentityService
}

func NewCryptoActionService(
	cryptoActionRepo cryptoActionRepo,
	identities identity.IdentityService,
) CryptoActionService {
	return CryptoActionService{
		cryptoActions:   cryptoActionRepo,
		identityService: identities,
	}
}

func (service CryptoActionService) ListCryptoActions(ctx context.Context, accountID string) ([]domain.CryptoAction, error) {
	return service.cryptoActions.List(ctx, accountID)
}

func (service CryptoActionService) CreateCryptoAction(ctx context.Context, actions []domain.CryptoAction) error {
	return service.cryptoActions.Create(ctx, actions)
}

func (service CryptoActionService) DeleteCryptoActionsUntil(ctx context.Context, accountID string, untilTime time.Time) error {
	return service.cryptoActions.DeleteUntil(ctx, accountID, untilTime)
}

func (service CryptoActionService) GetCryptoAction(ctx context.Context, actionID string) (domain.CryptoAction, error) {
	return service.cryptoActions.Get(ctx, actionID)
}

func (service CryptoActionService) CreateInvitationActions(ctx context.Context, senderID string, boxID string, identifierValue string, actionsDataJSON null.JSON) error {
	var actionsData map[string]string
	err := actionsDataJSON.Unmarshal(&actionsData)
	if err != nil {
		return merror.Transform(err).Describe("unmarshalling actions data")
	}

	identities, err := service.identityService.ListByIdentifier(ctx,
		domain.Identifier{
			Value: identifierValue,
			Kind:  domain.EmailIdentifier,
		},
	)
	if err != nil {
		return merror.Transform(err).Describe("retrieving identities")
	}

	if len(actionsData) != len(identities) {
		return merror.BadRequest().Describe(
			"required one entry per identity public key in for_server_no_store",
		)
	}

	actions := make([]domain.CryptoAction, len(identities))

	for i, identity := range identities {
		if !identity.Pubkey.Valid {
			return merror.Conflict().Describe("not all identities have a public key")
		}

		encryptedCryptoAction, present := actionsData[identity.Pubkey.String]
		if !present {
			return merror.BadRequest().Describef(
				"missing encrypted crypto action for pubkey \"%s\"",
				identity.Pubkey.String,
			)
		}

		action := domain.CryptoAction{
			AccountID:           identity.AccountID.String,
			Type:                "invitation",
			SenderIdentityID:    null.StringFrom(senderID),
			BoxID:               null.StringFrom(boxID),
			Encrypted:           encryptedCryptoAction,
			EncryptionPublicKey: identity.Pubkey.String,
			CreatedAt:           time.Now(),
		}
		actions[i] = action
	}

	return service.cryptoActions.Create(ctx, actions)
}
