package authz

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/format"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

// IsAMachine return true if the received claims correspond to a machine claims
// a machine authenticates itself as an sso client id using client_credentials so
// it checks if the subject equals the client to estimate it
func IsNotAMachine(ac oidc.AccessClaims) bool {
	return ac.ClientID != ac.Subject
}

// IsAMachine return the reverse of IsNotAMachine function
func IsAMachine(ac oidc.AccessClaims) bool {
	return !IsNotAMachine(ac)
}

func getBearerTokFromAuthorizationHeader(ctx echo.Context) (string, error) {
	// get authorization header
	header := ctx.Request().Header.Get("Authorization")
	if header == "" {
		return "", merr.NotFound().Ori(merr.OriHeaders)
	}

	// use format regexp to both check the token format and to retrieve it
	matches := format.BearerToken.FindStringSubmatch(header)
	if len(matches) != 2 {
		return "", merr.Unauthorized().Ori(merr.OriHeaders).Add("Authorization", merr.DVMalformed)
	}
	return matches[1], nil
}

func getBearerTokFromCookies(ctx echo.Context, tokenName, typeName string) (string, error) {
	// get authorization cookie
	cookieAccessToken, err := ctx.Request().Cookie(tokenName)
	if err != nil {
		return "", merr.NotFound().Ori(merr.OriCookies).Add(tokenName, merr.DVNotFound)
	}

	cookieTokenType, err := ctx.Request().Cookie(typeName)
	if err != nil || cookieTokenType.Value != "bearer" {
		return "", merr.Unauthorized().Desc(err.Error()).Ori(merr.OriCookies).Add(typeName, merr.DVInvalid)
	}

	if len(cookieAccessToken.Value) == 0 {
		return "", merr.Unauthorized().Desc(err.Error()).Ori(merr.OriCookies).Add(tokenName, merr.DVInvalid)
	}
	return cookieAccessToken.Value, nil
}

func SetCookie(ctx echo.Context, name, value string, duration time.Time) {
	// set for all cookies httponly, secure, samesitestrictmode...
	cookie := http.Cookie{
		HttpOnly: true, Secure: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}

	cookie.Name = name
	cookie.Value = value
	cookie.Expires = duration
	ctx.SetCookie(&cookie)
}

func DelCookies(ctx echo.Context, names ...string) {
	for _, name := range names {
		// set for all cookies httponly, secure, samesitestrictmode...
		cookie := http.Cookie{
			HttpOnly: true, Secure: true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		}

		cookie.Name = name
		cookie.Value = ""                               // empty the value
		cookie.Expires = time.Now().Add(-1 * time.Hour) // set expiration in the past
		cookie.MaxAge = -1                              // negative max age
		ctx.SetCookie(&cookie)
	}
}
