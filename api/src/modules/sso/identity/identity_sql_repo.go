package identity

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories/sqlboiler"
)

type identitySQLRepo struct {
	db *sql.DB
}

func NewIdentitySQLRepo(db *sql.DB) *identitySQLRepo {
	return &identitySQLRepo{
		db: db,
	}
}

func (repo *identitySQLRepo) Create(ctx context.Context, identity *Identity) error {
	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}

	identity.ID = id.String()
	// default value is minimal
	if identity.Notifications == "" {
		identity.Notifications = "minimal"
	}

	// convert to sql model
	return identity.toSQLBoiler().Insert(ctx, repo.db, boil.Infer())
}

func (repo *identitySQLRepo) Get(ctx context.Context, identityID string) (ret Identity, err error) {
	record, err := sqlboiler.FindIdentity(ctx, repo.db, identityID)
	if err == sql.ErrNoRows {
		return ret, merror.NotFound().Describe(err.Error()).Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return ret, err
	}
	return *ret.fromSQLBoiler(*record), nil
}

func (repo *identitySQLRepo) Update(ctx context.Context, identity *Identity) error {
	rowsAff, err := identity.toSQLBoiler().Update(ctx, repo.db, boil.Infer())
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Describe("no rows affected").Detail("id", merror.DVNotFound)
	}
	return nil
}

func (repo *identitySQLRepo) List(ctx context.Context, filters IdentityFilters) ([]*Identity, error) {
	mods := []qm.QueryMod{}
	if filters.IdentifierID.Valid {
		mods = append(mods, sqlboiler.IdentityWhere.IdentifierID.EQ(filters.IdentifierID.String))
	}
	if filters.IsAuthable.Valid {
		mods = append(mods, sqlboiler.IdentityWhere.IsAuthable.EQ(filters.IsAuthable.Bool))
	}
	if len(filters.IDs) > 0 {
		mods = append(mods, sqlboiler.IdentityWhere.ID.IN(filters.IDs))
	}
	if filters.AccountID.Valid {
		mods = append(mods, sqlboiler.IdentityWhere.AccountID.EQ(filters.AccountID))
	}

	// eager loading
	mods = append(mods, qm.Load("Identifier"))

	identityRecords, err := sqlboiler.Identities(mods...).All(ctx, repo.db)
	identities := make([]*Identity, len(identityRecords))
	if err == sql.ErrNoRows {
		return identities, nil
	}
	if err != nil {
		return identities, err
	}

	for i, record := range identityRecords {
		identities[i] = newIdentity().fromSQLBoiler(*record)
	}
	return identities, nil
}
