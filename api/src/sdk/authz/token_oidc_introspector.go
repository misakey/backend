package authz

import (
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester"
)

// NewOIDCIntrospector used to declare a route require OIDC authorization
// Information must be passed through a bearer token in Authorization HTTP Header
// The opaque token is instropected and information are set inside current context
// to be checked later by different actors (modules...)
func NewOIDCIntrospector(misakeyAudience string, selfRestrict bool, tokenRester rester.Client) echo.MiddlewareFunc {
	tokens := newOIDCIntroHTTP(misakeyAudience, tokenRester)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// handle bearer token
			opaqueTok, err := GetBearerTok(ctx)
			if err != nil {
				return err
			}

			// introspect the token and get the claims
			acc, err := tokens.GetClaims(ctx.Request().Context(), opaqueTok)
			if err != nil {
				if merror.HasCode(err, merror.InternalCode) {
					return merror.Transform(err).Describe("introspecting token")
				}
				return merror.Unauthorized().From(merror.OriHeaders).Describe(err.Error())
			}

			// only Misakey client can access our API routes
			if selfRestrict && acc.ClientID != misakeyAudience {
				return merror.Unauthorized().Describe("unauthorized client")
			}

			// set access claims in request context
			ctx.SetRequest(ctx.Request().WithContext(oidc.SetAccesses(ctx.Request().Context(), &acc)))
			return next(ctx)
		}
	}
}
