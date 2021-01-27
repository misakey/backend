package identity

import (
	"context"
	"database/sql"
	"encoding/base64"

	"github.com/duo-labs/webauthn/webauthn"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

type WebauthnIdentity struct {
	Identity

	WebAuthn []webauthn.Credential
}

func GetEmptyWebauthnIdentity(ctx context.Context, identity Identity) (WebauthnIdentity, error) {
	return WebauthnIdentity{
		identity,
		[]webauthn.Credential{},
	}, nil
}

func GetWebauthnIdentity(ctx context.Context, exec boil.ContextExecutor, identity Identity) (WebauthnIdentity, error) {
	// get credentials
	mods := []qm.QueryMod{
		sqlboiler.WebauthnCredentialWhere.IdentityID.EQ(identity.ID),
	}
	creds, err := sqlboiler.WebauthnCredentials(mods...).All(ctx, exec)
	if err != nil && err == sql.ErrNoRows {
		return WebauthnIdentity{}, merr.NotFound()
	}
	if err != nil {
		return WebauthnIdentity{}, err
	}

	wcreds := make([]webauthn.Credential, len(creds))

	for idx, cred := range creds {
		credID, err := base64.RawURLEncoding.DecodeString(cred.ID)
		if err != nil {
			return WebauthnIdentity{}, err
		}
		wcreds[idx] = webauthn.Credential{
			ID:              credID,
			PublicKey:       cred.PublicKey,
			AttestationType: cred.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:       cred.Aaguid,
				SignCount:    uint32(cred.SignCount),
				CloneWarning: cred.CloneWarning,
			},
		}
	}

	wid := WebauthnIdentity{
		identity,
		wcreds,
	}

	return wid, nil
}

func (wid *WebauthnIdentity) WebAuthnID() []byte {
	return []byte(wid.ID)
}

func (wid *WebauthnIdentity) WebAuthnName() string {
	return wid.DisplayName
}

func (wid *WebauthnIdentity) WebAuthnDisplayName() string {
	return wid.DisplayName
}

func (wid *WebauthnIdentity) WebAuthnIcon() string {
	return wid.AvatarURL.String
}

func (wid *WebauthnIdentity) WebAuthnCredentials() []webauthn.Credential {
	return wid.WebAuthn
}
