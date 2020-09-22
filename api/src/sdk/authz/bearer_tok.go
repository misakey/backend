package authz

import (
	"strings"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

func GetBearerTok(ctx echo.Context) (string, error) {
	// get authorization header
	var token string
	bearerTok := ctx.Request().Header.Get("Authorization")
	// TODO: use a real authentication process for websockets
	// and remove this part
	queryBearerTok := ctx.QueryParam("access_token")

	if len(bearerTok) == 0 {
		if len(queryBearerTok) == 0 {
			return "", merror.Unauthorized().From(merror.OriHeaders).
				Detail("Authorization", merror.DVRequired)
		} else {
			token = queryBearerTok
		}
	} else {
		// verify authorization header value is a bearer token and extract it
		tokSplit := strings.SplitAfter(bearerTok, "Bearer ")
		if len(tokSplit) != 2 {
			return "", merror.Unauthorized().
				From(merror.OriHeaders).
				Describef("token should be of form `Bearer {token}`").
				Detail("Authorization", merror.DVMalformed)
		}
		token = tokSplit[1]
	}

	return token, nil
}
