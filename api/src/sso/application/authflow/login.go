package authflow

import (
	"context"
	"fmt"
	"net/url"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow/login"
)

// GetLoginContext using a login challenge
func (afs Service) GetLoginContext(ctx context.Context, loginChallenge string) (login.Context, error) {
	return afs.authFlow.GetLoginContext(ctx, loginChallenge)
}

// BuildAndAcceptLogin takes the OIDCContext as the one used to login
// It builds the acceptance object and sends it as accepted to the authorization server
func (afs Service) BuildAndAcceptLogin(ctx context.Context, loginCtx login.Context) (string, error) {
	if len(loginCtx.OIDCContext.ACRValues()) == 0 {
		return "", fmt.Errorf("acr values are empty")
	}
	acr := loginCtx.OIDCContext.ACRValues().Get()
	acceptance := login.Acceptance{
		Subject: loginCtx.Subject,
		ACR:     acr,
		Context: loginCtx.OIDCContext,

		Remember:    (acr.RememberFor() > 0),
		RememberFor: acr.RememberFor(),
	}
	return afs.authFlow.Login(ctx, loginCtx.Challenge, acceptance)
}

// LoginRequiredErr helper
func (afs Service) LoginRequiredErr() string {
	return buildRedirectErr(merror.LoginRequiredCode, "forbidden prompt=none", afs.loginPageURL)
}

// LoginRedirectErr helper
func (afs Service) LoginRedirectErr(err error) string {
	return buildRedirectErr(merror.InvalidFlowCode, err.Error(), afs.loginPageURL)
}

// BuildLoginURL ...
func (afs Service) BuildLoginURL(loginChallenge string) string {
	// build the login URL
	finalURL := *afs.loginPageURL

	// add login_challenge to query params
	query := url.Values{}
	query.Set("login_challenge", loginChallenge)

	// escape query parameters
	finalURL.RawQuery = query.Encode()
	return finalURL.String()

}
