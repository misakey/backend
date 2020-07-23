package authz

import (
	"strings"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

func GetBearerTok(ctx echo.Context) (string, error) {
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
