package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories/sqlboiler"
)

type AccountSQLBoiler struct {
	db *sql.DB
}

func NewAccountSQLBoiler(db *sql.DB) *AccountSQLBoiler {
	return &AccountSQLBoiler{
		db: db,
	}
}

func (repo AccountSQLBoiler) toDomain(boilModel *sqlboiler.Account) *domain.Account {
	return &domain.Account{
		ID:            boilModel.ID,
		Password:      boilModel.Password,
		BackupData:    boilModel.BackupData,
		BackupVersion: boilModel.BackupVersion,
	}
}

func (repo AccountSQLBoiler) toSQLBoiler(domModel *domain.Account) *sqlboiler.Account {
	return &sqlboiler.Account{
		ID:            domModel.ID,
		Password:      domModel.Password,
		BackupData:    domModel.BackupData,
		BackupVersion: domModel.BackupVersion,
	}
}

func (repo AccountSQLBoiler) Create(ctx context.Context, account *domain.Account) error {
	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}

	account.ID = id.String()

	// convert domain to sql model
	sqlAccount := repo.toSQLBoiler(account)

	return sqlAccount.Insert(ctx, repo.db, boil.Infer())
}

func (repo AccountSQLBoiler) Get(ctx context.Context, accountID string) (ret domain.Account, err error) {
	sqlAccount, err := sqlboiler.FindAccount(ctx, repo.db, accountID)
	if err == sql.ErrNoRows {
		return ret, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return ret, err
	}

	return *repo.toDomain(sqlAccount), nil
}

func (repo AccountSQLBoiler) Update(ctx context.Context, account *domain.Account) error {
	sqlAccount := repo.toSQLBoiler(account)

	rowsAff, err := sqlAccount.Update(ctx, repo.db, boil.Infer())
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Detail("id", merror.DVNotFound).
			Describe("no account rows affected on udpate")
	}
	return nil
}
