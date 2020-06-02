package authz

import (
	"context"
	"fmt"
	"net/url"

	"gitlab.misakey.dev/misakey/msk-sdk-go/rester"
)

// tokenIntroHTTP implements Token repository interface using HTTP REST
type tokenIntroHTTP struct {
	formAdminRester rester.Client
}

// newTokenIntroHTTP is the tokenIntroHTTP structure constructor
func newTokenIntroHTTP(
	formAdminRester rester.Client,
) tokenIntroHTTP {
	return tokenIntroHTTP{
		formAdminRester: formAdminRester,
	}
}

func (h tokenIntroHTTP) Introspect(ctx context.Context, opaqueTok string) (instropection, error) {
	introTok := instropection{}
	route := fmt.Sprintf("/oauth2/introspect")

	params := url.Values{}
	params.Add("token", opaqueTok)

	err := h.formAdminRester.Post(ctx, route, nil, params, &introTok)
	return introTok, err
}
