package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type CryptoActionSQLBoiler struct {
	db *sql.DB
}

func NewCryptoActionSQLBoiler(db *sql.DB) *CryptoActionSQLBoiler {
	return &CryptoActionSQLBoiler{
		db: db,
	}
}

func (repo CryptoActionSQLBoiler) toDomain(boilModel *sqlboiler.CryptoAction) domain.CryptoAction {
	return domain.CryptoAction{
		ID:                  boilModel.ID,
		AccountID:           boilModel.AccountID,
		SenderIdentityID:    boilModel.SenderIdentityID,
		Type:                boilModel.Type,
		BoxID:               boilModel.BoxID,
		EncryptionPublicKey: boilModel.EncryptionPublicKey,
		Encrypted:           boilModel.Encrypted,
		CreatedAt:           boilModel.CreatedAt,
	}
}

func (repo CryptoActionSQLBoiler) toSQLBoiler(src domain.CryptoAction) *sqlboiler.CryptoAction {
	return &sqlboiler.CryptoAction{
		ID:                  src.ID,
		AccountID:           src.AccountID,
		SenderIdentityID:    src.SenderIdentityID,
		Type:                src.Type,
		BoxID:               src.BoxID,
		EncryptionPublicKey: src.EncryptionPublicKey,
		Encrypted:           src.Encrypted,
		CreatedAt:           src.CreatedAt,
	}
}

func (repo CryptoActionSQLBoiler) Create(ctx context.Context, actions []domain.CryptoAction) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	defer atomic.SQLRollback(ctx, tx, err)
	if err != nil {
		return err
	}

	for _, action := range actions {
		sqlAction := repo.toSQLBoiler(action)

		err = sqlAction.Insert(ctx, repo.db, boil.Infer())
		if err != nil {
			return merror.Transform(err).Describe("inserting action")
		}
	}

	return tx.Commit()
}

func (repo CryptoActionSQLBoiler) List(ctx context.Context, accountID string) ([]domain.CryptoAction, error) {
	records, err := sqlboiler.CryptoActions(
		sqlboiler.CryptoActionWhere.AccountID.EQ(accountID),
		qm.OrderBy(sqlboiler.CryptoActionColumns.CreatedAt+" ASC"),
	).All(ctx, repo.db)
	result := make([]domain.CryptoAction, len(records))
	if err == sql.ErrNoRows {
		return result, nil
	}
	if err != nil {
		return nil, err
	}

	for i, record := range records {
		result[i] = repo.toDomain(record)
	}
	return result, nil
}

func (repo CryptoActionSQLBoiler) Get(ctx context.Context, actionID string, accountID string) (domain.CryptoAction, error) {
	record, err := sqlboiler.CryptoActions(
		sqlboiler.CryptoActionWhere.ID.EQ(actionID),
		sqlboiler.CryptoActionWhere.AccountID.EQ(accountID),
	).One(ctx, repo.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.CryptoAction{}, merror.NotFound().Describef("no action with ID %s", actionID)
		}
		return domain.CryptoAction{}, err
	}

	return repo.toDomain(record), nil
}

func (repo CryptoActionSQLBoiler) DeleteUntil(ctx context.Context, accountID string, untilTime time.Time) error {
	rowsAff, err := sqlboiler.CryptoActions(
		sqlboiler.CryptoActionWhere.CreatedAt.LTE(untilTime),
		sqlboiler.CryptoActionWhere.AccountID.EQ(accountID),
	).DeleteAll(ctx, repo.db)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		// This should not happen
		// because the HTTP query is "delete until action with such ID"
		// and the action with this ID should be retrieved first by application layer
		// so it should exists, and at least this one should be deleted
		return merror.NotFound().Describe("no crypto actions to delete")
	}
	return nil
}

func (repo CryptoActionSQLBoiler) Delete(ctx context.Context, actionID string, accountID string) error {
	rowsAff, err := sqlboiler.CryptoActions(
		sqlboiler.CryptoActionWhere.ID.EQ(actionID),
		sqlboiler.CryptoActionWhere.AccountID.EQ(accountID),
	).DeleteAll(ctx, repo.db)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Describe("no crypto actions to delete")
	}
	return nil
}
