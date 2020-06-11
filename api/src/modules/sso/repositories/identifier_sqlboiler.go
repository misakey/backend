package repositories

import (
	"context"
	"database/sql"
	"strings"

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
		Value: strings.ToLower(identifier.Value),
		Kind:  string(identifier.Kind),
	}

	return sqlIdentifier.Insert(ctx, repo.db, boil.Infer())
}

func (repo *IdentifierSQLBoiler) Get(ctx context.Context, id string) (domain.Identifier, error) {
	mods := []qm.QueryMod{
		sqlboiler.IdentifierWhere.ID.EQ(id),
	}

	identifier, err := sqlboiler.Identifiers(mods...).One(ctx, repo.db)
	if err == sql.ErrNoRows {
		return domain.Identifier{}, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return domain.Identifier{}, err
	}

	return domain.Identifier{
		ID:    identifier.ID,
		Value: identifier.Value,
		Kind:  domain.IdentifierKind(identifier.Kind),
	}, nil
}

func (repo *IdentifierSQLBoiler) GetByKindValue(ctx context.Context, kind domain.IdentifierKind, value string) (domain.Identifier, error) {
	mods := []qm.QueryMod{
		sqlboiler.IdentifierWhere.Kind.EQ(string(kind)),
		sqlboiler.IdentifierWhere.Value.EQ(strings.ToLower(value)),
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
		Kind:  domain.IdentifierKind(identifier.Kind),
	}, nil
}
