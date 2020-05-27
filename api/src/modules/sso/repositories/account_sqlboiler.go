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

func (repo AccountSQLBoiler) Create(ctx context.Context, account *domain.Account) error {
	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}

	account.ID = id.String()

	// convert domain to sql model
	sqlAccount := sqlboiler.Account{
		ID:       account.ID,
		Password: account.Password,
	}

	return sqlAccount.Insert(ctx, repo.db, boil.Infer())
}

func (repo AccountSQLBoiler) Get(ctx context.Context, accountID string) (ret domain.Account, err error) {
	account, err := sqlboiler.FindAccount(ctx, repo.db, accountID)
	if err == sql.ErrNoRows {
		return ret, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return ret, err
	}

	ret.ID = account.ID
	ret.HasPassword = (len(account.Password) != 0)
	return ret, nil
}
