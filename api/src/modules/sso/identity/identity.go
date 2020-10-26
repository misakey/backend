package identity

import (
	"context"

	"github.com/volatiletech/null/v8"

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
	if err := ids.identities.Create(ctx, identity); err != nil {
		return merror.Transform(err).Describe("creating identity")
	}
	return nil
}

func (ids IdentityService) Get(ctx context.Context, identityID string) (ret Identity, err error) {
	if ret, err = ids.identities.Get(ctx, identityID); err != nil {
		return ret, merror.Transform(err).Describe("getting identity")
	}

	// retrieve the related identifier
	ret.Identifier, err = ids.identifierService.Get(ctx, ret.IdentifierID)
	if err != nil {
		return ret, merror.Transform(err).Describe("getting identifier")
	}
	return ret, nil
}

func (ids IdentityService) List(ctx context.Context, filters IdentityFilters) ([]*Identity, error) {
	return ids.identities.List(ctx, filters)
}

func (ids IdentityService) GetAuthableByIdentifierID(ctx context.Context, identifierID string) (Identity, error) {
	filters := IdentityFilters{
		IdentifierID: null.StringFrom(identifierID),
		IsAuthable:   null.BoolFrom(true),
	}
	identities, err := ids.identities.List(ctx, filters)
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
	if err := ids.identities.Update(ctx, identity); err != nil {
		return merror.Transform(err).Describe("updating identity")
	}
	return nil
}
