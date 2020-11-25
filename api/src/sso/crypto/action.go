package crypto

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
)

type cryptoActionRepo interface {
	Get(ctx context.Context, actionID string, accountID string) (Action, error)
	List(ctx context.Context, accountID string) ([]Action, error)
	Create(ctx context.Context, actions []Action) error
	DeleteUntil(ctx context.Context, accountID string, untilTime time.Time) error
	Delete(ctx context.Context, actionID string, accountID string) error
}

type CryptoActionService struct {
	cryptoActions cryptoActionRepo
}

func NewActionService(
	cryptoActionRepo cryptoActionRepo,
) CryptoActionService {
	return CryptoActionService{
		cryptoActions: cryptoActionRepo,
	}
}

//
// action models
//

type Action struct {
	ID                  string      `json:"id"`
	AccountID           string      `json:"-"`
	SenderIdentityID    null.String `json:"-"`
	Type                string      `json:"type"`
	BoxID               null.String `json:"box_id"`
	EncryptionPublicKey string      `json:"encryption_public_key"`
	Encrypted           string      `json:"encrypted"`
	CreatedAt           time.Time   `json:"created_at"`
}

func newAction() *Action { return &Action{} }

func (a *Action) fromSQLBoiler(boilModel sqlboiler.CryptoAction) *Action {
	a.ID = boilModel.ID
	a.AccountID = boilModel.AccountID
	a.SenderIdentityID = boilModel.SenderIdentityID
	a.Type = boilModel.Type
	a.BoxID = boilModel.BoxID
	a.EncryptionPublicKey = boilModel.EncryptionPublicKey
	a.Encrypted = boilModel.Encrypted
	a.CreatedAt = boilModel.CreatedAt
	return a
}

func (a Action) toSQLBoiler() *sqlboiler.CryptoAction {
	return &sqlboiler.CryptoAction{
		ID:                  a.ID,
		AccountID:           a.AccountID,
		SenderIdentityID:    a.SenderIdentityID,
		Type:                a.Type,
		BoxID:               a.BoxID,
		EncryptionPublicKey: a.EncryptionPublicKey,
		Encrypted:           a.Encrypted,
		CreatedAt:           a.CreatedAt,
	}
}

//
// action functions
//

func CreateActions(
	ctx context.Context, exec boil.ContextExecutor,
	actions []Action,
) error {
	for _, action := range actions {
		err := action.toSQLBoiler().Insert(ctx, exec, boil.Infer())
		if err != nil {
			return merror.Transform(err).Describe("inserting action")
		}
	}

	return nil

}

func CreateInvitationActions(
	ctx context.Context, exec boil.ContextExecutor,
	senderID, boxID, boxTitle, identifierValue string, actionsDataJSON null.JSON,
) error {
	var actionsData map[string]string
	err := actionsDataJSON.Unmarshal(&actionsData)
	if err != nil {
		return merror.Transform(err).Describe("unmarshalling actions data")
	}

	identities, err := identity.ListByIdentifier(ctx, exec,
		identity.Identifier{
			Value: identifierValue,
			Kind:  identity.EmailIdentifier,
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

	actions := make([]Action, len(identities))
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

		action := Action{
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

	if err := CreateActions(ctx, exec, actions); err != nil {
		return err
	}

	for identityID, notifDetailsBytes := range notifDetailsByIdentityID {
		err = identity.NotificationCreate(ctx, exec,
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

func GetAction(
	ctx context.Context, exec boil.ContextExecutor,
	actionID, accountID string,
) (Action, error) {
	record, err := sqlboiler.CryptoActions(
		sqlboiler.CryptoActionWhere.ID.EQ(actionID),
		sqlboiler.CryptoActionWhere.AccountID.EQ(accountID),
	).One(ctx, exec)
	if err != nil {
		if err == sql.ErrNoRows {
			return Action{}, merror.NotFound().Describef("no action with ID %s", actionID)
		}
		return Action{}, err
	}
	return *newAction().fromSQLBoiler(*record), nil
}

func ListActions(ctx context.Context, exec boil.ContextExecutor, accountID string) ([]Action, error) {
	records, err := sqlboiler.CryptoActions(
		sqlboiler.CryptoActionWhere.AccountID.EQ(accountID),
		qm.OrderBy(sqlboiler.CryptoActionColumns.CreatedAt+" ASC"),
	).All(ctx, exec)
	result := make([]Action, len(records))
	if err == sql.ErrNoRows {
		return result, nil
	}
	if err != nil {
		return nil, err
	}

	for i, record := range records {
		result[i] = *newAction().fromSQLBoiler(*record)
	}
	return result, nil
}

func DeleteAction(
	ctx context.Context, exec boil.ContextExecutor,
	actionID, accountID string,
) error {
	err := identity.NotificationMarkAutoInvitationUsed(ctx, exec, actionID)
	if err != nil {
		return merror.Transform(err).Describe(`marking related invitations as used`)
	}

	// TODO use the transaction here too
	rowsAff, err := sqlboiler.CryptoActions(
		sqlboiler.CryptoActionWhere.ID.EQ(actionID),
		sqlboiler.CryptoActionWhere.AccountID.EQ(accountID),
	).DeleteAll(ctx, exec)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Describe("no crypto actions to delete")
	}
	return nil
}
