package authflow

import (
	"context"
	"fmt"
	"net/url"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
	"gitlab.misakey.dev/misakey/msk-sdk-go/oauth"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

// LoginGetContext using a login challenge
func (afs AuthFlowService) GetLoginContext(ctx context.Context, loginChallenge string) (login.Context, error) {
	return afs.authFlow.GetLoginContext(ctx, loginChallenge)
}

// BuildAndAcceptLogin takes the OIDCContext as the one used to login
// It builds the acceptance object and sends it as accepted to the authorization server
func (afs AuthFlowService) BuildAndAcceptLogin(ctx context.Context, loginCtx login.Context) (login.Redirect, error) {
	if len(loginCtx.OIDCContext.ACRValues) == 0 {
		return login.Redirect{}, fmt.Errorf("acr values are empty")
	}
	acr := loginCtx.OIDCContext.ACRValues.Get()
	acceptance := login.Acceptance{
		Subject: loginCtx.Subject,
		ACR:     acr,
		Context: authn.NewContext().SetAMR(loginCtx.OIDCContext.AMRs),

		Remember:    (acr.RememberFor() > 0),
		RememberFor: acr.RememberFor(),
	}
	return afs.authFlow.Login(ctx, loginCtx.Challenge, acceptance)
}

// LoginRequiredErr helper
func (afs AuthFlowService) LoginRequiredErr() string {
	return oauth.BuildRedirectErr(merror.LoginRequiredCode, "forbidden prompt=none", afs.loginPageURL)
}

// LoginRedirectErr helper
func (afs AuthFlowService) LoginRedirectErr(err error) string {
	return oauth.BuildRedirectErr(merror.InvalidFlowCode, err.Error(), afs.loginPageURL)
}

// buildLoginURL
func (afs AuthFlowService) BuildLoginURL(loginChallenge string) string {
	// build the login URL
	finalURL := *afs.loginPageURL

	// add login_challenge to query params
	query := url.Values{}
	query.Set("login_challenge", loginChallenge)

	// escape query parameters
	finalURL.RawQuery = query.Encode()
	return finalURL.String()

}
