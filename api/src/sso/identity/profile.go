package identity

import (
	"context"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
	Contactable bool `json:"contactable"`
	NonIdentifiedPubkey null.String `json:"non_identified_pubkey"`
}
type ConfigProfileView struct {
	Email bool `json:"email"`
}

//
// Service profile related methods
//

func ProfileGet(ctx context.Context, exec boil.ContextExecutor, identityID string) (p ProfileView, err error) {
	// first retrieve the identity
	identity, err := Get(ctx, exec, identityID)
	if err != nil {
		return p, merror.Transform(err).Describe("getting identity")
	}
	// fill information considering profile configuration
	consents, err := listProfileSharingConsents(ctx, exec,
		profileSharingConsentFilters{
			revoked:    null.BoolFrom(false), // get active consents only
			identityID: null.StringFrom(identityID),
		},
	)
	if err != nil {
		return p, merror.Transform(err).Describe("getting profile sharing consent")
	}
	p.ID = identity.ID
	p.DisplayName = identity.DisplayName
	p.AvatarURL = identity.AvatarURL
	p.NonIdentifiedPubkey = identity.NonIdentifiedPubkey
	// for now only the email can be shared
	// NOTE: the shape/logic of the profile might change later with more information to hide/share
	for _, consent := range consents {
		if consent.informationType == string(identity.Identifier.Kind) {
			p.IdentifierID = identity.Identifier.ID
			p.Identifier.Value = identity.Identifier.Value
			p.Identifier.Kind = string(identity.Identifier.Kind)
		}
	}
	p.Contactable = true
	if identity.AccountID.IsZero() || !identity.NonIdentifiedPubkey.Valid {
		p.Contactable = false
	}

	return p, nil
}

func ProfileConfigShare(
	ctx context.Context, exec boil.ContextExecutor,
	identityID, informationType string,
) error {
	consent := profileSharingConsent{
		identityID:      identityID,
		informationType: informationType,
	}
	return createProfileSharingConsent(ctx, exec, &consent)
}

func ProfileConfigUnshare(
	ctx context.Context, exec boil.ContextExecutor,
	identityID, informationType string,
) error {
	return revokeConsentByIdentityType(ctx, exec, identityID, informationType)
}

func ProfileConfigGet(ctx context.Context, exec boil.ContextExecutor, identityID string) (c ConfigProfileView, err error) {
	// fill information considering profile configuration
	consents, err := listProfileSharingConsents(ctx, exec, profileSharingConsentFilters{
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
		case string(EmailIdentifier):
			c.Email = true
		}
	}
	return c, nil
}
