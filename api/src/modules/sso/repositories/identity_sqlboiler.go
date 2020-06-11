package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories/sqlboiler"
)

type IdentitySQLBoiler struct {
	db *sql.DB
}

func NewIdentitySQLBoiler(db *sql.DB) *IdentitySQLBoiler {
	return &IdentitySQLBoiler{
		db: db,
	}
}

func (repo IdentitySQLBoiler) toSqlBoiler(domModel *domain.Identity) *sqlboiler.Identity {
	return &sqlboiler.Identity{
		ID:            domModel.ID,
		AccountID:     domModel.AccountID,
		IdentifierID:  domModel.IdentifierID,
		IsAuthable:    domModel.IsAuthable,
		DisplayName:   domModel.DisplayName,
		Notifications: domModel.Notifications,
		AvatarURL:     domModel.AvatarURL,
		Confirmed:     domModel.Confirmed,
	}
}

func (repo IdentitySQLBoiler) toDomain(boilModel *sqlboiler.Identity) *domain.Identity {
	result := &domain.Identity{
		ID:            boilModel.ID,
		AccountID:     boilModel.AccountID,
		IdentifierID:  boilModel.IdentifierID,
		IsAuthable:    boilModel.IsAuthable,
		DisplayName:   boilModel.DisplayName,
		Notifications: boilModel.Notifications,
		AvatarURL:     boilModel.AvatarURL,
		Confirmed:     boilModel.Confirmed,
	}

	if boilModel.R != nil {
		identifier := boilModel.R.Identifier
		result.Identifier.ID = identifier.ID
		result.Identifier.Kind = domain.IdentifierKind(identifier.Kind)
		result.Identifier.Value = identifier.Value
	}

	return result
}

func (repo *IdentitySQLBoiler) Create(ctx context.Context, identity *domain.Identity) error {
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

	// convert domain to sql model
	sqlIdentity := repo.toSqlBoiler(identity)
	return sqlIdentity.Insert(ctx, repo.db, boil.Infer())
}

func (repo *IdentitySQLBoiler) Get(ctx context.Context, identityID string) (ret domain.Identity, err error) {
	identity, err := sqlboiler.FindIdentity(ctx, repo.db, identityID)
	if err == sql.ErrNoRows {
		return ret, merror.NotFound().Describe(err.Error()).Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return ret, err
	}
	return *repo.toDomain(identity), nil

}

func (repo *IdentitySQLBoiler) Update(ctx context.Context, identity *domain.Identity) error {
	rowsAff, err := repo.toSqlBoiler(identity).Update(ctx, repo.db, boil.Infer())
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Describe("no rows affected").Detail("id", merror.DVNotFound)
	}
	return nil
}

func (repo *IdentitySQLBoiler) Confirm(ctx context.Context, identityID string) error {
	// try to get identity
	identity, err := sqlboiler.Identities(sqlboiler.IdentityWhere.ID.EQ(identityID)).One(ctx, repo.db)
	if err != nil {
		return merror.NotFound().Detail("id", merror.DVNotFound)
	}

	identity.Confirmed = true

	rowsAff, err := identity.Update(ctx, repo.db, boil.Infer())
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Describe("could not find identity")
	}

	return nil
}

func (repo *IdentitySQLBoiler) List(ctx context.Context, filters domain.IdentityFilters) ([]*domain.Identity, error) {
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

	// eager loading
	mods = append(mods, qm.Load("Identifier"))

	identityRecords, err := sqlboiler.Identities(mods...).All(ctx, repo.db)
	domainIdentities := make([]*domain.Identity, len(identityRecords))
	if err == sql.ErrNoRows {
		return domainIdentities, nil
	}
	if err != nil {
		return domainIdentities, err
	}

	for i, record := range identityRecords {
		domainIdentities[i] = repo.toDomain(record)
	}
	return domainIdentities, nil
}
