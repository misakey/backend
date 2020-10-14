package oidc

import (
	"context"
	"net/http"
)

// BearerTokenAuthenticator is the default authenticator of the http client
type BearerTokenAuthenticator struct {
}

// Set an Authorization to Bearer + {jwt token} if the token is found in context
// Used to authorize intern calls between services
func (BearerTokenAuthenticator) Set(ctx context.Context, req *http.Request) {
	acc := GetAccesses(ctx)
	if acc != nil {
		req.Header.Set("Authorization", "Bearer "+acc.JWT)
	}
}
