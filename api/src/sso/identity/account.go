package identity

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type Account struct {
	ID            string
	Password      string
	BackupData    string
	BackupVersion int
}

func newAccount() *Account { return &Account{} }

func (a Account) toSQLBoiler() *sqlboiler.Account {
	return &sqlboiler.Account{
		ID:            a.ID,
		Password:      a.Password,
		BackupData:    a.BackupData,
		BackupVersion: a.BackupVersion,
	}
}

func (a *Account) fromSQLBoiler(boilModel *sqlboiler.Account) *Account {
	a.ID = boilModel.ID
	a.Password = boilModel.Password
	a.BackupData = boilModel.BackupData
	a.BackupVersion = boilModel.BackupVersion
	return a
}

func CreateAccount(ctx context.Context, exec boil.ContextExecutor, account *Account) error {
	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}

	account.ID = id.String()

	// convert domain to sql model and insert it
	return account.toSQLBoiler().Insert(ctx, exec, boil.Infer())
}

func GetAccount(ctx context.Context, exec boil.ContextExecutor, accountID string) (ret Account, err error) {
	sqlAccount, err := sqlboiler.FindAccount(ctx, exec, accountID)
	if err == sql.ErrNoRows {
		return ret, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return ret, err
	}

	return *newAccount().fromSQLBoiler(sqlAccount), nil
}

func UpdateAccount(ctx context.Context, exec boil.ContextExecutor, account *Account) error {
	rowsAff, err := account.toSQLBoiler().Update(ctx, exec, boil.Infer())
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Detail("id", merror.DVNotFound).
			Describe("no account rows affected on udpate")
	}
	return nil
}
