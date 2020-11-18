package cryptoaction

import (
	"context"
	"encoding/json"
	"time"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
)

type cryptoActionRepo interface {
	Get(ctx context.Context, actionID string, accountID string) (domain.CryptoAction, error)
	List(ctx context.Context, accountID string) ([]domain.CryptoAction, error)
	Create(ctx context.Context, actions []domain.CryptoAction) error
	DeleteUntil(ctx context.Context, accountID string, untilTime time.Time) error
	Delete(ctx context.Context, actionID string, accountID string) error
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

func (service CryptoActionService) DeleteCryptoAction(ctx context.Context, actionID string, accountID string) error {
	tx, err := service.identityService.SqlDB.BeginTx(ctx, nil)
	if err != nil {
		return merror.Transform(err).Describe(`creating transaction`)
	}
	err = service.identityService.NotificationMarkAutoInvitationUsed(ctx, tx, actionID)
	if err != nil {
		return merror.Transform(err).Describe(`marking related invitations as used`)
	}

	// TODO use the transaction here too
	err = service.cryptoActions.Delete(ctx, actionID, accountID)
	if err != nil {
		secondErr := tx.Rollback()
		if secondErr != nil {
			return merror.Transform(secondErr).Describef(`while handling error "%s"`, err.Error())
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return merror.Transform(err).Describe(`commiting notif transaction (cryptoaction was deleted though!)`)
	}
	return nil
}

func (service CryptoActionService) GetCryptoAction(ctx context.Context, actionID string, accountID string) (domain.CryptoAction, error) {
	return service.cryptoActions.Get(ctx, actionID, accountID)
}

func (service CryptoActionService) CreateInvitationActions(ctx context.Context, senderID string, boxID string, boxTitle string, identifierValue string, actionsDataJSON null.JSON) error {
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
	notifDetailsByIdentityID := make(map[string][]byte, len(identities))

	for i, identity := range identities {
		if !identity.Pubkey.Valid {
			return merror.Conflict().Describe("not all identities have a public key")
		}
		// cryptoaction

		encryptedCryptoAction, present := actionsData[identity.Pubkey.String]
		if !present {
			return merror.BadRequest().Describef(
				"missing encrypted crypto action for pubkey \"%s\"",
				identity.Pubkey.String,
			)
		}

		actionID, err := uuid.NewString()
		if err != nil {
			return merror.Transform(err).Describe("generating action UUID")
		}

		action := domain.CryptoAction{
			ID:                  actionID,
			AccountID:           identity.AccountID.String,
			Type:                "invitation",
			SenderIdentityID:    null.StringFrom(senderID),
			BoxID:               null.StringFrom(boxID),
			Encrypted:           encryptedCryptoAction,
			EncryptionPublicKey: identity.Pubkey.String,
			CreatedAt:           time.Now(),
		}
		actions[i] = action

		// notification (details differ for each identity)
		notifDetailsBytes, err := json.Marshal(struct {
			BoxID          string `json:"box_id"`
			BoxTitle       string `json:"box_title"`
			CryptoActionID string `json:"cryptoaction_id"`
		}{
			BoxID:          boxID,
			BoxTitle:       boxTitle,
			CryptoActionID: action.ID,
		})
		if err != nil {
			return merror.Transform(err).Describef(
				"marshalling notif details for pubkey \"%s\"",
				identity.Pubkey.String,
			)
		}
		notifDetailsByIdentityID[identity.ID] = notifDetailsBytes

	}

	err = service.cryptoActions.Create(ctx, actions)
	if err != nil {
		return err
	}

	for identityID, notifDetailsBytes := range notifDetailsByIdentityID {
		err = service.identityService.NotificationCreate(ctx,
			identityID,
			"box.auto_invite",
			null.JSONFrom(notifDetailsBytes),
		)
	}
	if err != nil {
		return err
	}

	return nil
}
