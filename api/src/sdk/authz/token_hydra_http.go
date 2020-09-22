package authz

import (
	"context"
	"net/url"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester"
)

// oidcIntroHTTP implements Token repository interface using HTTP REST
type oidcIntroHTTP struct {
	formAdminRester rester.Client
}

// newOIDCIntroHTTP is the oidcIntroHTTP structure constructor
func newOIDCIntroHTTP(
	formAdminRester rester.Client,
) oidcIntroHTTP {
	return oidcIntroHTTP{
		formAdminRester: formAdminRester,
	}
}

func (h oidcIntroHTTP) Introspect(ctx context.Context, opaqueTok string) (instropection, error) {
	introTok := instropection{}
	route := "/oauth2/introspect"

	params := url.Values{}
	params.Add("token", opaqueTok)

	err := h.formAdminRester.Post(ctx, route, nil, params, &introTok)
	return introTok, err
}
