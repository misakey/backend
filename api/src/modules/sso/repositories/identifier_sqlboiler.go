package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories/sqlboiler"
)

type IdentifierSQLBoiler struct {
	db *sql.DB
}

func NewIdentifierSQLBoiler(db *sql.DB) *IdentifierSQLBoiler {
	return &IdentifierSQLBoiler{
		db: db,
	}
}

func (repo *IdentifierSQLBoiler) Create(ctx context.Context, identifier *domain.Identifier) error {
	// generate new UUID for new record
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}
	identifier.ID = id.String()

	// convert domain to sql model
	sqlIdentifier := sqlboiler.Identifier{
		ID:    identifier.ID,
		Value: identifier.Value,
		Kind:  identifier.Kind,
	}

	fmt.Println(sqlIdentifier)
	return sqlIdentifier.Insert(ctx, repo.db, boil.Infer())
}

func (repo *IdentifierSQLBoiler) GetByKindValue(ctx context.Context, kind string, value string) (domain.Identifier, error) {
	mods := []qm.QueryMod{
		sqlboiler.IdentifierWhere.Kind.EQ(kind),
		sqlboiler.IdentifierWhere.Value.EQ(value),
	}

	identifier, err := sqlboiler.Identifiers(mods...).One(ctx, repo.db)
	if err == sql.ErrNoRows {
		return domain.Identifier{}, merror.NotFound().Detail("kind", merror.DVNotFound).Detail("value", merror.DVNotFound)
	}
	if err != nil {
		return domain.Identifier{}, err
	}

	return domain.Identifier{
		ID:    identifier.ID,
		Value: identifier.Value,
		Kind:  identifier.Kind,
	}, nil
}
