package echorouter

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// TokenLookup is a string in the form of "<source>:<name>" that is used
// to extract token from the request.
// Possible values:
// - "header:<name>"
// - "query:<name>"
// - "cookie:<name>"
const echoTokenLookupRule = "header:Authorization"

// JWTContextKey is the key where the jwt.Token will be set in echo Context
const echoJWTContextKey = "jwt"

// NewJWTMidlw return an echo middleware function which requires a valid
// Access JWT as an Authorization header.
// It raises an error if no JWT is provided or invalid, according to its strict parameter.
func NewJWTMidlw(strict bool) echo.MiddlewareFunc {
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte(ajwt.JWTStaticSignature),
		TokenLookup: echoTokenLookupRule,
		ContextKey:  echoJWTContextKey,
		Claims:      &ajwt.AccessClaims{},
		// skip JWT's check if not strict and Authorization header is empty
		Skipper: func(ctx echo.Context) bool {
			return !strict && len(ctx.Request().Header.Get("Authorization")) == 0
		},
		SuccessHandler: func(ctx echo.Context) {
			jwToken := ctx.Get(echoJWTContextKey).(*jwt.Token)
			claims := jwToken.Claims.(*ajwt.AccessClaims)
			// set raw JWT inside our custom data structure so it can be forwarded to other services
			claims.SetRawJWT(jwToken.Raw)

			// we prefer to set our custom claim data structure directly inside the request context
			accessCtx := ajwt.SetAccesses(ctx.Request().Context(), claims)
			ctx.SetRequest(ctx.Request().WithContext(accessCtx))
		},
		// return 401 in all cases (sometimes the middleware returns 400 if no jwt is forwarded)
		ErrorHandler: func(err error) error {
			return merror.Unauthorized().
				Describef("missing, invalid or expired jwt: %s", err.Error()).
				From(merror.OriHeaders).
				Detail("Authorization", merror.DVInvalid)
		},
	})
}
