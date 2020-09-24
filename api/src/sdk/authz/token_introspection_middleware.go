package authz

import (
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester"
)

// NewOIDCIntrospector used to declare a route require OIDC authorization
// Information must be passed through a bearer token in Authorization HTTP Header
// The opaque token is instropected and information are set inside current context
// to be checked later by different actors (modules...)
func NewOIDCIntrospector(misakeyAudience string, selfRestrict bool, tokenRester rester.Client) echo.MiddlewareFunc {
	tokens := newOIDCIntroHTTP(tokenRester)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// handle bearer token
			opaqueTok, err := GetBearerTok(ctx)
			if err != nil {
				return err
			}

			// introspect the token
			introTok, err := tokens.Introspect(ctx.Request().Context(), opaqueTok)
			if err != nil {
				if merror.HasCode(err, merror.InternalCode) {
					return merror.Transform(err).Describe("introspecting token")
				}
				return merror.Unauthorized().From(merror.OriHeaders).Describe(err.Error())
			}

			// check access token is active
			// see https://www.ory.sh/docs/hydra/sdk/api#oauth2tokenintrospection to know what is an active token
			if !introTok.Active {
				return merror.Unauthorized().
					From(merror.OriHeaders).
					Describe("token is not active").
					Detail("Authorization", merror.DVInvalid)
			}

			// check token is an access_token
			if introTok.TokenType != "access_token" {
				return merror.Unauthorized().From(merror.OriHeaders).
					Describe("token must be an access token").
					Detail("Authorization", merror.DVInvalid)
			}

			// only Misakey client can access our API routes
			if selfRestrict && introTok.ClientID != misakeyAudience {
				return merror.Unauthorized().Describe("unauthorized client")
			}

			// fill a claim structure with introspection
			ac := ajwt.AccessClaims{
				Issuer: introTok.Issuer,
				// access token aren't usable externally
				Audiences: []string{misakeyAudience},
				ClientID:  introTok.ClientID,

				ExpiresAt: introTok.ExpiresAt,
				IssuedAt:  introTok.IssuedAt,
				NotBefore: introTok.NotBefore,

				Subject:    introTok.Subject,
				IdentityID: introTok.Ext.MID,
				AccountID:  null.StringFrom(introTok.Ext.AID),

				Scope: introTok.Scope,
				ACR:   introTok.Ext.ACR,

				Token: opaqueTok,
			}

			// valid the access claim
			if err := ac.Valid(); err != nil {
				return err
			}

			// set access claims in request context
			ctx.SetRequest(ctx.Request().WithContext(ajwt.SetAccesses(ctx.Request().Context(), &ac)))
			return next(ctx)
		}
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
		ACR ajwt.ACRSecLvl `json:"acr"`
		MID string         `json:"mid"`
		AID string         `json:"aid"`
	} `json:"ext"`
}
