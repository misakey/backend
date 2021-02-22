package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

type Client struct {
	ID string `json:"client_id"`

	Name    string `json:"client_name"`
	LogoURI string `json:"logo_uri"`

	Scope         string   `json:"scope"`
	GrantTypes    []string `json:"grant_types"`
	RedirectURIs  []string `json:"redirect_uris"`
	ResponseTypes []string `json:"response_types"`

	Audience           []string `json:"audience"`
	AllowedCorsOrigins []string `json:"allowed_cors_origins"`

	SubjectType               string `json:"subject_type"`
	UserinfoSignedResponseALG string `json:"userinfo_signed_response_ald"`
	TokenEndpointAuthMethod   string `json:"token_endpoint_auth_method"`
	Secret                    string `json:"client_secret"`
	SecretExpiresAt           int    `json:"client_secret_expires_at"`

	// Note handled nowadays, to add when client authenticates using private_key_jwt
	// JWKs Jwks `json:"jwks"`
}

// type Jwks struct {
// 	Keys []Keys `json:"keys"`
// }

// type Keys struct {
// 	Alg string   `json:"alg"`
// 	Crv string   `json:"crv"`
// 	D   string   `json:"d"`
// 	Dp  string   `json:"dp"`
// 	Dq  string   `json:"dq"`
// 	E   string   `json:"e"`
// 	K   string   `json:"k"`
// 	Kid string   `json:"kid"`
// 	Kty string   `json:"kty"`
// 	N   string   `json:"n"`
// 	P   string   `json:"p"`
// 	Q   string   `json:"q"`
// 	Qi  string   `json:"qi"`
// 	Use string   `json:"use"`
// 	X   string   `json:"x"`
// 	X5C []string `json:"x5c"`
// 	Y   string   `json:"y"`
// }

func (afs Service) UpdateClientSecret(ctx context.Context, cliID string, newSecret string) error {
	cli, err := afs.authFlow.GetClient(ctx, cliID)
	// ignore not found because client will be created if not existing yet
	if err != nil && !merr.IsANotFound(err) {
		return merr.From(err).Desc("getting client")
	}
	if merr.IsANotFound(err) {
		cli.ID = cliID
		cli.Name = cliID
		cli.Scope = "openid"
		cli.Audience = []string{afs.homePageURL.String(), afs.selfCliID}
		cli.GrantTypes = []string{"client_credentials"}
		cli.ResponseTypes = []string{"token"}
		cli.SubjectType = "pairwise"
		cli.UserinfoSignedResponseALG = "none"
		cli.TokenEndpointAuthMethod = "client_secret_post"
		cli.Secret = newSecret
		cli.SecretExpiresAt = 0
		if err := afs.authFlow.CreateClient(ctx, &cli); err != nil {
			return merr.From(err).Desc("creating client")
		}
		return nil
	}
	cli.Secret = newSecret
	if err := afs.authFlow.UpdateClient(ctx, &cli); err != nil {
		return merr.From(err).Desc("updating client")
	}
	return nil
}
