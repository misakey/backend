package application

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/mwebauthn"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

// BeginWebAuthnRegistrationQuery ...
type BeginWebAuthnRegistrationQuery struct {
	identityID string
}

// BindAndValidate ...
func (cmd *BeginWebAuthnRegistrationQuery) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	cmd.identityID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
	)
}

// BeginWebAuthnRegistration returns options to register webauthn credentials
func (sso *SSOService) BeginWebAuthnRegistration(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*BeginWebAuthnRegistrationQuery)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}
	if acc.IdentityID != query.identityID {
		return nil, merr.Forbidden()
	}

	curIdentity, err := identity.Get(ctx, sso.ssoDB, query.identityID)
	if err != nil {
		return nil, merr.From(err).Desc("getting identityID")
	}

	wid, err := identity.GetWebauthnIdentity(ctx, sso.ssoDB, curIdentity)
	if err != nil {
		return nil, merr.From(err).Desc("creating webauthn identity")
	}

	// get current credentials to avoid duplicates
	excludeCredentials := make([]protocol.CredentialDescriptor, len(wid.WebAuthnCredentials()))

	for i, credential := range wid.WebAuthnCredentials() {
		var credentialDescriptor protocol.CredentialDescriptor
		credentialDescriptor.CredentialID = credential.ID
		credentialDescriptor.Type = protocol.PublicKeyCredentialType
		excludeCredentials[i] = credentialDescriptor
	}

	options, sessionData, err := sso.AuthenticationService.WebauthnHandler.BeginRegistration(&wid, webauthn.WithExclusions(excludeCredentials))
	if err != nil {
		return nil, merr.From(err).Desc("beginning webauthn registration")
	}

	if err := mwebauthn.StoreSession(sso.redConn, sessionData, curIdentity.ID, options.Response.Challenge.String()); err != nil {
		return nil, merr.From(err).Desc("storing session")
	}

	return options, nil
}

// CredentialsView only takes some of the credentials attributes
// to expose them via the API
type CredentialsView struct {
	ID         string    `json:"id"`
	IdentityID string    `json:"identity_id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}

// FinishWebAuthnRegistrationQuery ...
type FinishWebAuthnRegistrationQuery struct {
	identityID    string
	RawCredential string `json:"credential"`
	credential    *protocol.ParsedCredentialCreationData
	Name          string `json:"name"`
}

// BindAndValidate ...
func (cmd *FinishWebAuthnRegistrationQuery) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	var err error
	cmd.credential, err = protocol.ParseCredentialCreationResponseBody(strings.NewReader(cmd.RawCredential))
	if err != nil {
		return merr.From(err).Ori(merr.OriBody)
	}
	cmd.identityID = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.identityID, v.Required, is.UUIDv4),
		v.Field(&cmd.Name, v.Required, v.Length(1, 100)),
	)
}

// FinishWebAuthnRegistration records a webauthn credential
func (sso *SSOService) FinishWebAuthnRegistration(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*FinishWebAuthnRegistrationQuery)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}
	if acc.IdentityID != query.identityID {
		return nil, merr.Forbidden()
	}

	curIdentity, err := identity.Get(ctx, sso.ssoDB, query.identityID)
	if err != nil {
		return nil, merr.From(err).Desc("getting identityID")
	}

	sessionData, err := mwebauthn.GetSession(sso.redConn, curIdentity.ID, query.credential.Response.CollectedClientData.Challenge)
	if err != nil {
		return nil, merr.From(err).Desc("getting session")
	}

	wid, err := identity.GetWebauthnIdentity(ctx, sso.ssoDB, curIdentity)
	if err != nil {
		return nil, merr.From(err).Desc("creating webauthn identity")
	}

	cred, err := sso.AuthenticationService.WebauthnHandler.CreateCredential(&wid, sessionData, query.credential)
	if err != nil {
		return nil, merr.From(err).Desc("validating login")
	}

	credID := base64.RawURLEncoding.EncodeToString(cred.ID)
	toStore := sqlboiler.WebauthnCredential{
		ID:              credID,
		Name:            query.Name,
		IdentityID:      query.identityID,
		PublicKey:       cred.PublicKey,
		AttestationType: cred.AttestationType,
		Aaguid:          cred.Authenticator.AAGUID,
		SignCount:       int(cred.Authenticator.SignCount),
		CloneWarning:    cred.Authenticator.CloneWarning,
	}
	if err := toStore.Insert(ctx, sso.ssoDB, boil.Infer()); err != nil {
		return nil, merr.From(err).Desc("inserting credential")
	}

	return CredentialsView{
		ID:         toStore.ID,
		IdentityID: toStore.IdentityID,
		Name:       toStore.Name,
		CreatedAt:  toStore.CreatedAt,
	}, nil
}

// ListCredentialsQuery ...
type ListCredentialsQuery struct {
	IdentityID string `query:"identity_id" json:"-"`
}

// BindAndValidate ...
func (cmd *ListCredentialsQuery) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merr.From(err).Ori(merr.OriQuery)
	}

	return v.ValidateStruct(cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4),
	)
}

// ListCredentials for a given identity
func (sso *SSOService) ListCredentials(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*ListCredentialsQuery)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}
	if acc.IdentityID != query.IdentityID {
		return nil, merr.Forbidden()
	}

	mods := []qm.QueryMod{
		sqlboiler.WebauthnCredentialWhere.IdentityID.EQ(query.IdentityID),
	}

	credentials, err := sqlboiler.WebauthnCredentials(mods...).All(ctx, sso.ssoDB)
	if err == nil && credentials == nil {
		return []sqlboiler.WebauthnCredential{}, nil
	}
	if err != nil {
		return nil, merr.From(err).Desc("listing credential")
	}

	result := make([]CredentialsView, len(credentials))
	for i, cred := range credentials {
		result[i] = CredentialsView{
			ID:         cred.ID,
			IdentityID: cred.IdentityID,
			Name:       cred.Name,
			CreatedAt:  cred.CreatedAt,
		}
	}

	return result, nil
}

// DeleteCredentialQuery ...
type DeleteCredentialQuery struct {
	id string
}

// BindAndValidate ...
func (cmd *DeleteCredentialQuery) BindAndValidate(eCtx echo.Context) error {
	cmd.id = eCtx.Param("id")

	return v.ValidateStruct(cmd,
		v.Field(&cmd.id, v.Required),
	)
}

// DeleteCredential after checking it is owned by the requester
func (sso *SSOService) DeleteCredential(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*DeleteCredentialQuery)

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Forbidden()
	}

	curIdentity, err := identity.Get(ctx, sso.ssoDB, acc.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("getting identity")
	}

	cred, err := sqlboiler.FindWebauthnCredential(ctx, sso.ssoDB, query.id)
	if err != nil {
		return nil, merr.From(err).Desc("getting credential")
	}

	if cred.IdentityID != acc.IdentityID {
		return nil, merr.Forbidden()
	}

	number, err := mwebauthn.CredentialsNumber(ctx, sso.ssoDB, acc.IdentityID)
	if err != nil {
		return nil, merr.From(err).Desc("counting credentials")
	}

	if curIdentity.MFAMethod == "webauthn" && number <= 1 {
		return nil, merr.Forbidden().Desc("2FA configured and last credential remaining")
	}

	if _, err := cred.Delete(ctx, sso.ssoDB); err != nil {
		return nil, merr.From(err).Desc("deleting credential")
	}

	return nil, nil
}
