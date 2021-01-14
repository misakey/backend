package identity

import (
	"context"
	"database/sql"

	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

// Identity ...
type Identity struct {
	ID                  string         `json:"id"`
	AccountID           null.String    `json:"account_id"`
	IdentifierValue     string         `json:"identifier_value"`
	IdentifierKind      IdentifierKind `json:"identifier_kind"`
	DisplayName         string         `json:"display_name"`
	Notifications       string         `json:"notifications"`
	AvatarURL           null.String    `json:"avatar_url"`
	Color               null.String    `json:"color"`
	Level               int            `json:"level"`
	Pubkey              null.String    `json:"pubkey"`
	NonIdentifiedPubkey null.String    `json:"non_identified_pubkey"`
}

// IdentifierKind ...
type IdentifierKind string

const (
	// EmailIdentifier ...
	EmailIdentifier IdentifierKind = "email"
)

// Filters ...
type Filters struct {
	IdentifierValue null.String
	IDs             []string
	AccountID       null.String
}

func newIdentity() *Identity { return &Identity{} }

func (i Identity) toSQLBoiler() *sqlboiler.Identity {
	return &sqlboiler.Identity{
		ID:                  i.ID,
		AccountID:           i.AccountID,
		IdentifierValue:     i.IdentifierValue,
		IdentifierKind:      string(i.IdentifierKind),
		DisplayName:         i.DisplayName,
		Notifications:       i.Notifications,
		AvatarURL:           i.AvatarURL,
		Color:               i.Color,
		Level:               i.Level,
		Pubkey:              i.Pubkey,
		NonIdentifiedPubkey: i.NonIdentifiedPubkey,
	}
}

func (i *Identity) fromSQLBoiler(src sqlboiler.Identity) *Identity {
	i.ID = src.ID
	i.AccountID = src.AccountID
	i.IdentifierValue = src.IdentifierValue
	i.IdentifierKind = IdentifierKind(src.IdentifierKind)
	i.DisplayName = src.DisplayName
	i.Notifications = src.Notifications
	i.AvatarURL = src.AvatarURL
	i.Color = src.Color
	i.Level = src.Level
	i.Pubkey = src.Pubkey
	i.NonIdentifiedPubkey = src.NonIdentifiedPubkey
	return i
}

// Create ...
func Create(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, identity *Identity) error {
	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return merr.From(err).Desc("could not generate uuid v4")
	}

	identity.ID = id.String()
	// default value is minimal
	if identity.Notifications == "" {
		identity.Notifications = "minimal"
	}

	// convert to sql model
	if err := identity.toSQLBoiler().Insert(ctx, exec, boil.Infer()); err != nil {
		return err
	}

	// send notification message
	// for onboarding purpose
	if err := NotificationCreate(ctx, exec, redConn, identity.ID, "user.create_identity", null.JSONFromPtr(nil)); err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("notifying identity %s", identity.ID)
	}
	return nil
}

// Get ...
func Get(ctx context.Context, exec boil.ContextExecutor, identityID string) (ret Identity, err error) {
	mods := []qm.QueryMod{
		sqlboiler.IdentityWhere.ID.EQ(identityID),
	}
	record, err := sqlboiler.Identities(mods...).One(ctx, exec)
	if err == sql.ErrNoRows {
		return ret, merr.NotFound().Desc(err.Error()).Add("id", merr.DVNotFound)
	}
	if err != nil {
		return ret, err
	}
	return *ret.fromSQLBoiler(*record), nil
}

// GetByIdentifierValue...
func GetByIdentifierValue(ctx context.Context, exec boil.ContextExecutor, identifierValue string) (ret Identity, err error) {
	mods := []qm.QueryMod{
		sqlboiler.IdentityWhere.IdentifierValue.EQ(identifierValue),
	}
	record, err := sqlboiler.Identities(mods...).One(ctx, exec)
	if err == sql.ErrNoRows {
		return ret, merr.NotFound().Desc(err.Error()).Add("identifier_value", merr.DVNotFound)
	}
	if err != nil {
		return ret, err
	}
	return *ret.fromSQLBoiler(*record), nil
}

// List ...
func List(ctx context.Context, exec boil.ContextExecutor, filters Filters) ([]*Identity, error) {
	mods := []qm.QueryMod{}
	if len(filters.IDs) > 0 {
		mods = append(mods, sqlboiler.IdentityWhere.ID.IN(filters.IDs))
	}
	if filters.AccountID.Valid {
		mods = append(mods, sqlboiler.IdentityWhere.AccountID.EQ(filters.AccountID))
	}

	identityRecords, err := sqlboiler.Identities(mods...).All(ctx, exec)
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

// Update ...
func Update(ctx context.Context, exec boil.ContextExecutor, identity *Identity) error {
	rowsAff, err := identity.toSQLBoiler().Update(ctx, exec, boil.Infer())
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merr.NotFound().Desc("no rows affected").Add("id", merr.DVNotFound)
	}
	return nil
}
