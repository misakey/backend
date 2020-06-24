package authflow

import (
	"context"
	"net/url"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/consent"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
	"gitlab.misakey.dev/misakey/msk-sdk-go/oauth"
)

// ConsentGetContext
func (afs AuthFlowService) ConsentGetContext(ctx context.Context, consentChallenge string) (consent.Context, error) {
	// get info about current consent flow
	return afs.authFlow.GetConsentContext(ctx, consentChallenge)
}

// ConsentGetSessions
func (afs AuthFlowService) ConsentGetSessions(ctx context.Context, identityID string) ([]consent.Session, error) {
	return afs.authFlow.GetConsentSessions(ctx, identityID)
}

// ConsentAccept
func (afs AuthFlowService) ConsentAccept(ctx context.Context, consentChallenge string, acceptance consent.Acceptance) string {
	redirect, err := afs.authFlow.Consent(ctx, consentChallenge, acceptance)
	if err != nil {
		return oauth.BuildRedirectErr(merror.InvalidFlowCode, err.Error(), afs.consentPageURL)
	}
	return redirect.To
}

// ConsentRedirectErr helper
func (afs AuthFlowService) ConsentRedirectErr(err error) string {
	return oauth.BuildRedirectErr(merror.InvalidFlowCode, err.Error(), afs.consentPageURL)
}

// buildConsentURL
func (afs AuthFlowService) BuildConsentURL(consentChallenge string) string {
	// build the authentication URL
	authenticationURL, err := url.ParseRequestURI(afs.consentPageURL)
	if err != nil {
		return afs.ConsentRedirectErr(err)
	}

	// add consent_challenge to query params
	query := url.Values{}
	query.Set("consent_challenge", consentChallenge)

	// escape query parameters
	authenticationURL.RawQuery = query.Encode()
	return authenticationURL.String()

}
