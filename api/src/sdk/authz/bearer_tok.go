package authz

import (
	"strings"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// GetBearerTokFromCookie ...
func GetBearerTokFromCookie(ctx echo.Context) (string, error) {
	// get authorization cookie
	bearerTok, err := ctx.Request().Cookie("accesstoken")
	if err != nil {
		return "", merr.Unauthorized().Ori(merr.OriCookies).
			Add("accesstoken", merr.DVInvalid)
	}

	if len(bearerTok.Value) == 0 {
		return "", merr.Unauthorized().Ori(merr.OriCookies).
			Add("accesstoken", merr.DVRequired)
	}

	tokType, err := ctx.Request().Cookie("tokentype")
	if err != nil || tokType.Value != "bearer" {
		return "", merr.Unauthorized().Ori(merr.OriCookies).
			Add("tokentype", merr.DVInvalid)
	}

	return bearerTok.Value, nil
}

// GetBearerTokFromHeader ...
func GetBearerTokFromHeader(ctx echo.Context) (string, error) {
	// get authorization header
	bearerTok := ctx.Request().Header.Get("Authorization")

	if len(bearerTok) == 0 {
		return "", merr.Unauthorized().Ori(merr.OriHeaders).
			Add("Authorization", merr.DVRequired)
	}
	// verify authorization header value is a bearer token and extract it
	tokSplit := strings.SplitAfter(bearerTok, "Bearer ")
	if len(tokSplit) != 2 {
		return "", merr.Unauthorized().Ori(merr.OriHeaders).
			Descf("token should be of form `Bearer {token}`").
			Add("Authorization", merr.DVMalformed)
	}
	return tokSplit[1], nil
}
