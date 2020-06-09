package application

import (
	"context"
	"strings"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/consent"
)

// Init a user consent stage (a.k.a. consent flow)
// It interacts with hydra to know either user has already consented to share data with the RP
// It returns a URL user's agent should be redirected to
// Today, it accept directly the consent of the user with the email claim contained in the ID token
// TODO: handle CGU/TOS acceptance
// TODO: handle identity (identifier) choice ?
func (sso SSOService) ConsentInit(ctx context.Context, consentChallenge string) string {
	// 1. get consent context
	consentCtx, err := sso.authFlowService.ConsentGetContext(ctx, consentChallenge)
	if err != nil {
		return sso.authFlowService.ConsentRedirectErr(err)
	}

	// 2. retrieve subject information to put it in ID tokens as claims
	identity, err := sso.identityService.Get(ctx, consentCtx.Subject)
	if err != nil {
		return sso.authFlowService.ConsentRedirectErr(err)
	}
	identifier, err := sso.identifierService.Get(ctx, identity.IdentifierID)
	if err != nil {
		return sso.authFlowService.ConsentRedirectErr(err)
	}

	// 3. build consent accept request directly
	acceptance := consent.Acceptance{
		GrantScope:  consentCtx.RequestedScope, // accept all requested scopes
		Remember:    true,
		RememberFor: 0, // remember for ever the user consent
	}
	acceptance.Session.IDTokenClaims.Scope = strings.Join(consentCtx.RequestedScope, " ")
	acceptance.Session.IDTokenClaims.Email = identifier.Value
	acceptance.Session.IDTokenClaims.AMR = consentCtx.AuthnContext.GetAMR()
	acceptance.Session.AccessTokenClaims.ACR = consentCtx.ACR

	// 4. tell hydra the consent contract
	return sso.authFlowService.ConsentAccept(ctx, consentChallenge, acceptance)
}
