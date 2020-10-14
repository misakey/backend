package authz

import (
	"strings"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

func GetBearerTokFromCookie(ctx echo.Context) (string, error) {
	// get authorization cookie
	bearerTok, err := ctx.Request().Cookie("accesstoken")
	if err != nil {
		return "", merror.Unauthorized().From(merror.OriCookies).
			Detail("accesstoken", merror.DVInvalid)
	}

	if len(bearerTok.Value) == 0 {
		return "", merror.Unauthorized().From(merror.OriCookies).
			Detail("accesstoken", merror.DVRequired)
	}

	tokType, err := ctx.Request().Cookie("tokentype")
	if err != nil || tokType.Value != "bearer" {
		return "", merror.Unauthorized().From(merror.OriCookies).
			Detail("tokentype", merror.DVInvalid)
	}

	return bearerTok.Value, nil
}

func GetBearerTokFromHeader(ctx echo.Context) (string, error) {
	// get authorization header
	bearerTok := ctx.Request().Header.Get("Authorization")

	if len(bearerTok) == 0 {
		return "", merror.Unauthorized().From(merror.OriHeaders).
			Detail("Authorization", merror.DVRequired)
	}
	// verify authorization header value is a bearer token and extract it
	tokSplit := strings.SplitAfter(bearerTok, "Bearer ")
	if len(tokSplit) != 2 {
		return "", merror.Unauthorized().
			From(merror.OriHeaders).
			Describef("token should be of form `Bearer {token}`").
			Detail("Authorization", merror.DVMalformed)
	}
	return tokSplit[1], nil
}
