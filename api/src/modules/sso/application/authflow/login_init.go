package authflow

import (
	"context"
	"net/url"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
	"gitlab.misakey.dev/misakey/msk-sdk-go/oauth"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

// Init a user authentication stage (a.k.a. login flow)
// It interacts with hydra to know either user is already authenticated or not
// It returns a URL user's agent should be redirected to
func (afs AuthFlowService) LoginInit(ctx context.Context, loginChallenge string) string {
	// get info about current login session
	loginCtx, err := afs.authFlow.GetLoginContext(ctx, loginChallenge)
	if err != nil {
		return oauth.BuildRedirectErr(merror.InvalidFlowCode, err.Error(), afs.loginPageURL)
	}

	// TODO: handle session for identity ID corresponding to same accounts
	// skip indicates if an active session has been detected so the authentication is not required
	if loginCtx.Skip {
		acceptance := login.Acceptance{
			Subject: loginCtx.Subject,
		}
		successRedirect, err := afs.authFlow.Login(ctx, loginCtx.Challenge, acceptance)
		if err != nil {
			return oauth.BuildRedirectErr(merror.InvalidFlowCode, err.Error(), afs.loginPageURL)
		}
		return successRedirect.To
	}

	// build the authentication URL
	authenticationURL, err := url.ParseRequestURI(afs.loginPageURL)
	if err != nil {
		return oauth.BuildRedirectErr(merror.InvalidURLCode, err.Error(), afs.loginPageURL)
	}

	// add login_challenge to query params
	query := url.Values{}
	query.Set("login_challenge", loginCtx.Challenge)

	// escape query parameters
	authenticationURL.RawQuery = query.Encode()
	return authenticationURL.String()
}
