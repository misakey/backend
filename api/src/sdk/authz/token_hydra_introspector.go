package authz

import (
	"context"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester"
)

// hydraInstropector implements Token repository interface using HYDRA HTTP REST API
type hydraInstropector struct {
	formAdminRester rester.Client

	audience    string
	selfCliOnly bool
}

// NewHydraIntrospector is the hydraInstropector structure constructor
// it expects as tokenRepo a rester.Client with form format and hydra admin url configured
func newHydraIntrospector(
	tokenRepo interface{},
	audience string, selfCliOnly bool,
) (h hydraInstropector, err error) {
	formAdminRester, ok := tokenRepo.(rester.Client)
	if !ok {
		return h, merr.Internal().Desc("expect a rester.Client as tokenRepo")
	}
	return hydraInstropector{
		formAdminRester: formAdminRester,
		audience:        audience,
		selfCliOnly:     selfCliOnly,
	}, nil
}

// hydraInstropection from Hydra
type hydraInstropection struct {
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
	Active   bool   `json:"active"`
	TokenUse string `json:"token_use"`

	// Additional custom claims
	Ext struct {
		ACRValues oidc.ClassRefs `json:"acr_values"`
		MID       string         `json:"mid"`
		AID       string         `json:"aid"`
	} `json:"ext"`
}

// GetBearerTok using cookies and bearer token if not found in cookies...
func (h hydraInstropector) GetBearerTok(ctx echo.Context) (string, bool, error) {
	// try to get it from header and cookies
	headTok, headErr := getBearerTokFromAuthorizationHeader(ctx)
	cookTok, cookErr := getBearerTokFromCookies(ctx, "accesstoken", "tokentype")
	if headErr == nil && cookErr == nil {
		return "", false, merr.Unauthorized().Desc("too much way of giving authorization")
	}
	if headErr == nil {
		return headTok, false, nil
	}
	if cookErr == nil {
		return cookTok, true, nil
	}
	return "", false, merr.Unauthorized().Desc("missing authorization")
}

// GetClaims using hydra http introspection endpoint
func (h hydraInstropector) GetClaims(ctx context.Context, opaqueTok string) (ac oidc.AccessClaims, err error) {
	introTok := hydraInstropection{}
	route := "/oauth2/introspect"

	params := url.Values{}
	params.Add("token", opaqueTok)

	if err := h.formAdminRester.Post(ctx, route, nil, params, &introTok); err != nil {
		return ac, err
	}

	// check access token is active
	// see https://www.ory.sh/docs/hydra/sdk/api#oauth2tokenintrospection to know what is an active token
	if !introTok.Active {
		return ac, merr.Unauthorized().Ori(merr.OriHeaders).Add("Authorization", merr.DVInvalid).
			Desc("token is not active")
	}

	// check token is an access_token
	if introTok.TokenUse != "access_token" {
		return ac, merr.Unauthorized().Ori(merr.OriHeaders).Add("Authorization", merr.DVInvalid).
			Desc("token must be an access token")
	}

	// fill a claim structure with introspection
	ac = oidc.AccessClaims{
		Issuer: introTok.Issuer,
		// access token aren't usable externally
		Audiences: []string{h.audience}, // WHY is it forced here?
		// Audiences: introTok.Audiences,
		ClientID:  introTok.ClientID,
		ExpiresAt: introTok.ExpiresAt,
		IssuedAt:  introTok.IssuedAt,
		NotBefore: introTok.NotBefore,
		Subject:   introTok.Subject,
		Scope:     introTok.Scope,
		Token:     opaqueTok,
	}
	// if client_id and subject are the same, it means a client_credentials flow has been performed
	// and the requested is a machine (organizations)
	// fill claims according to this
	if introTok.ClientID == introTok.Subject {
		ac.IdentityID = introTok.Subject
		ac.ACR = oidc.ACR2 // client_credentials = acr 2
	} else { // other, it is an end-user
		ac.IdentityID = introTok.Ext.MID
		// use NewString to potentially set Valid to false on empty account ID
		ac.AccountID = null.NewString(introTok.Ext.AID, len(introTok.Ext.AID) > 0)
		ac.ACR = introTok.Ext.ACRValues.Get()

	}

	// valid claims
	if err := ac.Valid(); err != nil {
		return ac, err
	}
	return ac, nil
}

func (h hydraInstropector) CheckClientID(ctx context.Context, ac oidc.AccessClaims) error {
	if h.selfCliOnly && ac.ClientID != h.audience {
		return merr.Unauthorized().Ori(merr.OriCookies).Desc("unauthorized client")
	}
	return nil
}

// HandleErr
func (h hydraInstropector) HandleErr(_ echo.Context, _ echo.HandlerFunc, err error) error {
	return merr.From(err).Code(merr.UnauthorizedCode)
}
