package authz

import (
	"context"
	"net/url"

	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester"
)

// oidcIntroHTTP implements Token repository interface using HTTP REST
type oidcIntroHTTP struct {
	audience        string
	formAdminRester rester.Client
}

// newOIDCIntroHTTP is the oidcIntroHTTP structure constructor
func newOIDCIntroHTTP(
	audience string,
	formAdminRester rester.Client,
) oidcIntroHTTP {
	return oidcIntroHTTP{
		audience:        audience,
		formAdminRester: formAdminRester,
	}
}

// instropection from Hydra
type instropection struct {
	Audiences []string `json:"aud"`
	ClientID  string   `json:"client_id"`
	Scope     string   `json:"scope"`
	Subject   string   `json:"sub"`

	// fields validated by Hydra
	Issuer    string `json:"iss"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`

	// fields validated inside this service
	Active    bool   `json:"active"`
	TokenType string `json:"token_type"`

	// Additional custom claims
	Ext struct {
		ACRValues oidc.ClassRefs `json:"acr_values"`
		MID       string         `json:"mid"`
		AID       string         `json:"aid"`
	} `json:"ext"`
}

func (h oidcIntroHTTP) GetClaims(ctx context.Context, opaqueTok string) (ac oidc.AccessClaims, err error) {
	introTok := instropection{}
	route := "/oauth2/introspect"

	params := url.Values{}
	params.Add("token", opaqueTok)

	if err := h.formAdminRester.Post(ctx, route, nil, params, &introTok); err != nil {
		return ac, err
	}

	// check access token is active
	// see https://www.ory.sh/docs/hydra/sdk/api#oauth2tokenintrospection to know what is an active token
	if !introTok.Active {
		return ac, merror.Unauthorized().
			From(merror.OriHeaders).
			Describe("token is not active").
			Detail("Authorization", merror.DVInvalid)
	}

	// check token is an access_token
	if introTok.TokenType != "access_token" {
		return ac, merror.Unauthorized().From(merror.OriHeaders).
			Describe("token must be an access token").
			Detail("Authorization", merror.DVInvalid)
	}

	// fill a claim structure with introspection
	ac = oidc.AccessClaims{
		Issuer: introTok.Issuer,
		// access token aren't usable externally
		Audiences: []string{h.audience},
		ClientID:  introTok.ClientID,

		ExpiresAt: introTok.ExpiresAt,
		IssuedAt:  introTok.IssuedAt,
		NotBefore: introTok.NotBefore,

		Subject:    introTok.Subject,
		IdentityID: introTok.Ext.MID,
		// use NewString to potentially set Valid to false on empty account ID (on acr < 2)
		AccountID: null.NewString(introTok.Ext.AID, len(introTok.Ext.AID) > 0),

		Scope: introTok.Scope,
		ACR:   introTok.Ext.ACRValues.Get(),

		Token: opaqueTok,
	}
	return ac, ac.Valid()
}
