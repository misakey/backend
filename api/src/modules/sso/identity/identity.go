package identity

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type Identity struct {
	ID            string      `json:"id"`
	AccountID     null.String `json:"account_id"`
	IdentifierID  string      `json:"identifier_id"`
	IsAuthable    bool        `json:"is_authable"`
	DisplayName   string      `json:"display_name"`
	Notifications string      `json:"notifications"`
	AvatarURL     null.String `json:"avatar_url"`
	Color         null.String `json:"color"`
	Level         int         `json:"level"`
	Pubkey        null.String `json:"pubkey"`

	// Identifier is always returned within the identity entity as a nested JSON object
	Identifier domain.Identifier `json:"identifier"`
}

type IdentityFilters struct {
	IdentifierID null.String
	IsAuthable   null.Bool
	IDs          []string
	AccountID    null.String
}

func newIdentity() *Identity { return &Identity{} }

func (i Identity) toSQLBoiler() *sqlboiler.Identity {
	return &sqlboiler.Identity{
		ID:            i.ID,
		AccountID:     i.AccountID,
		IdentifierID:  i.IdentifierID,
		IsAuthable:    i.IsAuthable,
		DisplayName:   i.DisplayName,
		Notifications: i.Notifications,
		AvatarURL:     i.AvatarURL,
		Color:         i.Color,
		Level:         i.Level,
		Pubkey:        i.Pubkey,
	}
}

func (i *Identity) fromSQLBoiler(src sqlboiler.Identity) *Identity {
	i.ID = src.ID
	i.AccountID = src.AccountID
	i.IdentifierID = src.IdentifierID
	i.IsAuthable = src.IsAuthable
	i.DisplayName = src.DisplayName
	i.Notifications = src.Notifications
	i.AvatarURL = src.AvatarURL
	i.Color = src.Color
	i.Level = src.Level
	i.Pubkey = src.Pubkey

	if src.R != nil {
		identifier := src.R.Identifier
		i.Identifier.ID = identifier.ID
		i.Identifier.Kind = domain.IdentifierKind(identifier.Kind)
		i.Identifier.Value = identifier.Value
	}
	return i
}

//
// Service identity related methods
//

func (ids IdentityService) Create(ctx context.Context, identity *Identity) error {
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
	return identity.toSQLBoiler().Insert(ctx, ids.sqlDB, boil.Infer())
}

func (ids IdentityService) Get(ctx context.Context, identityID string) (ret Identity, err error) {
	record, err := sqlboiler.FindIdentity(ctx, ids.sqlDB, identityID)
	if err == sql.ErrNoRows {
		return ret, merror.NotFound().Describe(err.Error()).Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return ret, err
	}
	ret.fromSQLBoiler(*record)

	// retrieve the related identifier
	ret.Identifier, err = ids.identifierService.Get(ctx, ret.IdentifierID)
	if err != nil {
		return ret, merror.Transform(err).Describe("getting identifier")
	}
	return ret, nil
}

func (ids IdentityService) List(ctx context.Context, filters IdentityFilters) ([]*Identity, error) {
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

	identityRecords, err := sqlboiler.Identities(mods...).All(ctx, ids.sqlDB)
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

func (ids IdentityService) GetAuthableByIdentifierID(ctx context.Context, identifierID string) (Identity, error) {
	filters := IdentityFilters{
		IdentifierID: null.StringFrom(identifierID),
		IsAuthable:   null.BoolFrom(true),
	}
	identities, err := ids.List(ctx, filters)
	if err != nil {
		return Identity{}, err
	}
	if len(identities) < 1 {
		return Identity{}, merror.NotFound().
			Detail("identifier_id", merror.DVNotFound).
			Detail("is_authable", merror.DVNotFound)
	}
	if len(identities) > 1 {
		return Identity{}, merror.Internal().Describef("more than one authable identity found for %s", identifierID)
	}
	return *identities[0], nil
}

func (ids IdentityService) Update(ctx context.Context, identity *Identity) error {
	rowsAff, err := identity.toSQLBoiler().Update(ctx, ids.sqlDB, boil.Infer())
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Describe("no rows affected").Detail("id", merror.DVNotFound)
	}
	return nil
}

func (ids IdentityService) ListByIdentifier(ctx context.Context, identifier domain.Identifier) ([]*Identity, error) {
	identifier, err := ids.identifierService.GetByKindValue(ctx, identifier)
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving identifier")
	}

	return ids.List(ctx,
		IdentityFilters{
			IdentifierID: null.StringFrom(identifier.ID),
		})
}
