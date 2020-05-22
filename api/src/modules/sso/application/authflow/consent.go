package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/consent"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
	"gitlab.misakey.dev/misakey/msk-sdk-go/oauth"
)

// ConsentGetContext
func (afs AuthFlowService) ConsentGetContext(ctx context.Context, consentChallenge string) (consent.Context, error) {
	// get info about current consent flow
	return afs.authFlow.GetConsentContext(ctx, consentChallenge)
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
