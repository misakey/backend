package identity

import (
	"context"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

//
// models
//

type ProfileView struct {
	ID           string      `json:"id"`
	DisplayName  string      `json:"display_name"`
	AvatarURL    null.String `json:"avatar_url"`
	IdentifierID string      `json:"identifier_id"`
	Identifier   struct {
		Value string `json:"value"`
		Kind  string `json:"kind"`
	} `json:"identifier"`
}
type ConfigProfileView struct {
	Email bool `json:"email"`
}

type profileSharingConsent struct {
	id              int
	identityID      string
	informationType string
	createdAt       time.Time
	revokedAt       null.Time
}

func newProfileSharingConsent() *profileSharingConsent { return &profileSharingConsent{} }

//
// Service profile related methods
//

func (ids IdentityService) ProfileGet(ctx context.Context, identityID string) (p ProfileView, err error) {
	// first retrieve the identity
	identity, err := ids.Get(ctx, identityID)
	if err != nil {
		return p, merror.Transform(err).Describe("getting identity")
	}
	// fill information considering profile configuration
	consents, err := ids.profileSharingConsents.List(ctx, profileSharingConsentFilters{
		revoked:    null.BoolFrom(false), // get active consents only
		identityID: null.StringFrom(identityID),
	})
	if err != nil {
		return p, merror.Transform(err).Describe("getting profile sharing consent")
	}
	p.ID = identity.ID
	p.DisplayName = identity.DisplayName
	p.AvatarURL = identity.AvatarURL
	// for now only the email can be shared
	// NOTE: the shape/logic of the profile might change later with more information to hide/share
	for _, consent := range consents {
		if consent.informationType == string(identity.Identifier.Kind) {
			p.IdentifierID = identity.Identifier.ID
			p.Identifier.Value = identity.Identifier.Value
			p.Identifier.Kind = string(identity.Identifier.Kind)
		}
	}
	return p, nil
}

func (ids IdentityService) ProfileConfigShare(ctx context.Context, identityID, informationType string) error {
	consent := profileSharingConsent{
		identityID:      identityID,
		informationType: informationType,
	}
	return ids.profileSharingConsents.Create(ctx, &consent)
}

func (ids IdentityService) ProfileConfigUnshare(ctx context.Context, identityID, informationType string) error {
	return ids.profileSharingConsents.revokeByIdentityType(ctx, identityID, informationType)
}

func (ids IdentityService) ProfileConfigGet(ctx context.Context, identityID string) (c ConfigProfileView, err error) {
	// fill information considering profile configuration
	consents, err := ids.profileSharingConsents.List(ctx, profileSharingConsentFilters{
		revoked:    null.BoolFrom(false), // get active consents only
		identityID: null.StringFrom(identityID),
	})
	if err != nil {
		return c, merror.Transform(err).Describe("getting profile sharing consent")
	}
	// for now only the email can be shared
	// NOTE: the shape/logic of the profile might change later with more information to hide/share
	for _, consent := range consents {
		switch consent.informationType {
		case string(domain.EmailIdentifier):
			c.Email = true
		}
	}
	return c, nil
}
