package crypto

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/uuid"
)

// Action models and helpers
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

// CreateActions inserts the cryptoaction in DB.
// if the cryptoaction has not ID it will create one
func CreateActions(
	ctx context.Context, exec boil.ContextExecutor,
	actions []Action,
) error {
	for _, action := range actions {
		if action.ID == "" {
			actionID, err := uuid.NewString()
			if err != nil {
				return merr.From(err).Desc("generating action UUID")
			}

			action.ID = actionID
		}

		err := action.toSQLBoiler().Insert(ctx, exec, boil.Infer())
		if err != nil {
			return merr.From(err).Desc("inserting action")
		}
	}

	return nil

}

// GetAction ...
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
			return Action{}, merr.NotFound().Descf("no action with ID %s", actionID)
		}
		return Action{}, err
	}
	return *newAction().fromSQLBoiler(*record), nil
}

// ListActions ...
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

// DeleteAction ...
func DeleteAction(
	ctx context.Context, exec boil.ContextExecutor,
	actionID, accountID string,
) error {
	err := identity.NotificationMarkAutoInvitationUsed(ctx, exec, actionID)
	if err != nil {
		return merr.From(err).Desc(`marking related invitations as used`)
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
		return merr.NotFound().Desc("no crypto actions to delete")
	}
	return nil
}
