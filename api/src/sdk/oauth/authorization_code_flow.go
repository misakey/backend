package oauth

import "gitlab.misakey.dev/misakey/backend/api/src/sdk/rester"

// AuthorizationCodeFlow (from OpenID Connect) using a private_key_jwt method for the final token exchange
// more info: https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowSteps
// allows to perform easily two steps:
// - the build of the url to ask for an authorization code
// - the exchange of the code vs final tokens
type AuthorizationCodeFlow struct {
	// clientID as OAuth2 Client ID
	clientID string

	// codeURL as the auth URL to request the authorization code.
	codeURL string
	// redirectCodeURL as the URL where the Auth server should redirect to with the authorization code
	redirectCodeURL string

	// token URL to inform as private_key_jwt audience
	tokenURL string
	// redirectTokenURL as the URL where this service should redirect to with the access token
	redirectTokenURL string

	tokenRester rester.Client
}

// NewAuthorizationCodeFlow is AuthorizationCodeFlow's constructor
func NewAuthorizationCodeFlow(
	cliID string,
	codeURL string, redirectCodeURL string,
	tokenRester rester.Client, tokenURL string, redirectTokenURL string,
) (*AuthorizationCodeFlow, error) {
	return &AuthorizationCodeFlow{
		clientID: cliID,
		codeURL:  codeURL, redirectCodeURL: redirectCodeURL,
		tokenRester: tokenRester, tokenURL: tokenURL, redirectTokenURL: redirectTokenURL,
	}, nil
}
