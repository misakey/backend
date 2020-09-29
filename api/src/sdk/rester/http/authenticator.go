package http

import (
	"context"
	"net/http"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
)

type authenticator interface {
	Set(context.Context, *http.Request)
}

// BearerTokenAuthenticator is the default authenticator of the http client
type BearerTokenAuthenticator struct {
}

// Set an Authorization to Bearer + {jwt token} if the token is found in context
// Used to authorize intern calls between services
func (_ *BearerTokenAuthenticator) Set(ctx context.Context, req *http.Request) {
	acc := ajwt.GetAccesses(ctx)
	if acc != nil {
		req.Header.Set("Authorization", "Bearer "+acc.JWT)
	}
}