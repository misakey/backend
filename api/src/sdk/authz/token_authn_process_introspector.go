package authz

import (
	"context"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

type processRepo interface {
	GetClaims(ctx context.Context, token string) (oidc.AccessClaims, error)
}

// NewAuthnProcessIntrospector ...
// Authorization must be passed through a bearer token in Authorization HTTP Header
// The opaque token is instropected and information are set inside current context
// to be checked later by different actors (modules...)
func NewAuthnProcessIntrospector(selfCliID string, tokens processRepo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// handle bearer token - ignore invalid header
			opaqueTok, err := GetBearerTokFromHeader(ctx)
			if err != nil {
				return next(ctx)
			}

			// introspect the token and retrieve access claims
			ac, err := tokens.GetClaims(ctx.Request().Context(), opaqueTok)
			if err != nil {
				// not found authorization is possible during an authn process
				// some authn step will be required in this case
				// we don't raise any error regarding this
				if merror.HasCode(err, merror.NotFoundCode) {
					return next(ctx)
				}
				return merror.Unauthorized().From(merror.OriHeaders).Describe(err.Error())
			}

			// set access claims in request context
			ctx.SetRequest(ctx.Request().WithContext(oidc.SetAccesses(ctx.Request().Context(), &ac)))
			return next(ctx)
		}
	}
}
