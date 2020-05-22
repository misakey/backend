package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
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

func (repo *IdentitySQLBoiler) Create(ctx context.Context, identity *domain.Identity) error {
	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}

	identity.ID = id.String()

	// convert domain to sql model
	sqlIdentity := sqlboiler.Identity{
		ID:            identity.ID,
		AccountID:     identity.AccountID,
		IdentifierID:  identity.IdentifierID,
		IsAuthable:    identity.IsAuthable,
		DisplayName:   identity.DisplayName,
		Notifications: identity.Notifications,
		AvatarURL:     null.StringFrom(identity.AvatarURL),
		Confirmed:     identity.Confirmed,
	}

	return sqlIdentity.Insert(ctx, repo.db, boil.Infer())
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
