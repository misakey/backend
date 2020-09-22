package authn

import (
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// NewProcessIntrospector
// Authorization must be passed through a bearer token in Authorization HTTP Header
// The opaque token is instropected and information are set inside current context
// to be checked later by different actors (modules...)
func NewProcessIntrospector(selfCliID string, tokens processRepo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// handle bearer token - ignore invalid header
			opaqueTok, err := authz.GetBearerTok(ctx)
			if err != nil {
				return next(ctx)
			}

			// introspect the token
			process, err := tokens.GetByTok(ctx.Request().Context(), opaqueTok)
			if err != nil {
				// not found authorization is possible during an authn process
				// some authn step will be required in this case
				// we don't raise any error regarding this
				if merror.HasCode(err, merror.NotFoundCode) {
					return next(ctx)
				}
				return merror.Unauthorized().From(merror.OriHeaders).Describe(err.Error())
			}

			// fill a claim structure with introspection
			ac := ajwt.AccessClaims{
				Issuer: selfCliID,
				// access token aren't usable externally
				Audiences: []string{selfCliID},
				ClientID:  selfCliID,

				ExpiresAt: process.ExpiresAt,
				IssuedAt:  process.IssuedAt,
				NotBefore: process.IssuedAt,

				Subject:    process.LoginChallenge,
				ACR:        ajwt.ACRSecLvl(process.CompleteAMRs.ToACR()),
				IdentityID: process.IdentityID, // potentially empty

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
