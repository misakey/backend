package identity

import (
	"context"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// ProfileView ...
type ProfileView struct {
	ID                        string      `json:"id"`
	DisplayName               string      `json:"display_name"`
	AvatarURL                 null.String `json:"avatar_url"`
	IdentifierValue           string      `json:"identifier_value"`
	IdentifierKind            string      `json:"identifier_kind"`
	Contactable               bool        `json:"contactable"`
	NonIdentifiedPubkey       null.String `json:"non_identified_pubkey"`
	NonIdentifiedPubkeyAesRsa null.String `json:"non_identified_pubkey_aes_rsa"`
}

// ConfigProfileView ...
type ConfigProfileView struct {
	Email bool `json:"email"`
}

// ProfileGet ...
func ProfileGet(ctx context.Context, exec boil.ContextExecutor, identityID string) (p ProfileView, err error) {
	// first retrieve the identity
	identity, err := Get(ctx, exec, identityID)
	if err != nil {
		return p, merr.From(err).Desc("getting identity")
	}
	// fill information considering profile configuration
	consents, err := listProfileSharingConsents(ctx, exec,
		profileSharingConsentFilters{
			revoked:    null.BoolFrom(false), // get active consents only
			identityID: null.StringFrom(identityID),
		},
	)
	if err != nil {
		return p, merr.From(err).Desc("getting profile sharing consent")
	}
	p.ID = identity.ID
	p.DisplayName = identity.DisplayName
	p.AvatarURL = identity.AvatarURL
	p.NonIdentifiedPubkey = identity.NonIdentifiedPubkey
	p.NonIdentifiedPubkeyAesRsa = identity.NonIdentifiedPubkeyAesRsa
	// for now only the email can be shared
	// NOTE: the shape/logic of the profile might change later with more information to hide/share
	for _, consent := range consents {
		if consent.informationType == "email" {
			p.IdentifierValue = identity.IdentifierValue
			p.IdentifierKind = "email"
			break
		}
	}
	p.Contactable = true
	if identity.AccountID.IsZero() || !identity.NonIdentifiedPubkey.Valid {
		p.Contactable = false
	}

	return p, nil
}

// ProfileConfigShare ...
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

// ProfileConfigUnshare ...
func ProfileConfigUnshare(
	ctx context.Context, exec boil.ContextExecutor,
	identityID, informationType string,
) error {
	return revokeConsentByIdentityType(ctx, exec, identityID, informationType)
}

// ProfileConfigGet ...
func ProfileConfigGet(ctx context.Context, exec boil.ContextExecutor, identityID string) (c ConfigProfileView, err error) {
	// fill information considering profile configuration
	consents, err := listProfileSharingConsents(ctx, exec, profileSharingConsentFilters{
		revoked:    null.BoolFrom(false), // get active consents only
		identityID: null.StringFrom(identityID),
	})
	if err != nil {
		return c, merr.From(err).Desc("getting profile sharing consent")
	}
	// for now only the email can be shared
	// NOTE: the shape/logic of the profile might change later with more information to hide/share
	for _, consent := range consents {
		switch consent.informationType {
		case "email":
			c.Email = true
		}
	}
	return c, nil
}
