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
// It returns a URL user's agent should be redirect to
func (h *Handler) InitLogin(ctx context.Context, challenge string) string {
	// get info about current login session
	loginCtx, err := h.authFlow.GetLoginContext(ctx, challenge)
	if err != nil {
		return oauth.BuildRedirectErr(merror.InvalidFlowCode, err.Error(), h.loginPageURL)
	}

	// skip indicates if an active session has been detected so the authentication is not required
	if loginCtx.Skip {
		acceptance := login.Acceptance{
			Subject: loginCtx.Subject,
		}
		successRedirect, err := h.authFlow.Login(ctx, challenge, acceptance)
		if err != nil {
			return oauth.BuildRedirectErr(merror.InvalidFlowCode, err.Error(), h.loginPageURL)
		}
		return successRedirect.To
	}

	// build the authentication URL
	authenticationURL, err := url.ParseRequestURI(h.loginPageURL)
	if err != nil {
		return oauth.BuildRedirectErr(merror.InvalidURLCode, err.Error(), h.loginPageURL)
	}

	// add login_challenge to query params
	query := url.Values{}
	query.Set("login_challenge", challenge)

	// escape query parameters
	authenticationURL.RawQuery = query.Encode()
	return authenticationURL.String()
}
