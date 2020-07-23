package authflow

import (
	"context"
	"net/url"
	"strings"

	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/slice"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/consent"
)

// GetConsentContext
func (afs AuthFlowService) GetConsentContext(ctx context.Context, consentChallenge string) (consent.Context, error) {
	// get info about current consent flow
	return afs.authFlow.GetConsentContext(ctx, consentChallenge)
}

// BuildAndAcceptConsent takes the RequestedScope as consented.
// It builds the acceptance object and sends it as accepted to the authorization server
func (afs AuthFlowService) BuildAndAcceptConsent(
	ctx context.Context,
	consentCtx consent.Context,
	identifierValue string,
) string {
	acceptance := consent.Acceptance{
		GrantScope:  consentCtx.RequestedScope, // accept all requested scopes
		Remember:    true,
		RememberFor: 0, // remember for ever the user consent
	}
	acceptance.Session.IDTokenClaims.Scope = strings.Join(consentCtx.RequestedScope, " ")
	acceptance.Session.IDTokenClaims.Email = identifierValue
	acceptance.Session.IDTokenClaims.AMR = consentCtx.OIDCContext.AMRs()
	// Add IdentityID and Account ID in ID Token for the Misakey Application to be able to use it
	// External RPs are not able to access this information so the involved ClientID is checked
	if consentCtx.Client.ID == afs.selfCliID {
		acceptance.Session.IDTokenClaims.MID = null.StringFrom(consentCtx.OIDCContext.MID())
		acceptance.Session.IDTokenClaims.AID = consentCtx.OIDCContext.AID()
	}
	acceptance.Session.AccessTokenClaims = consentCtx.OIDCContext
	redirect, err := afs.authFlow.Consent(ctx, consentCtx.Challenge, acceptance)
	if err != nil {
		return buildRedirectErr(merror.InvalidFlowCode, err.Error(), afs.consentPageURL)
	}
	return redirect.To
}

// ShouldSkipConsent returns a boolean corresponding to Skipable and
// a potential error that may occur during the computation of the boolean.

// the ssoClientID (currently involved client) is used to check if
// the implicit consent is allowed (the other identities' consent linked to the account make the consent automatic)
func (afs AuthFlowService) ShouldSkipConsent(
	ctx context.Context,
	requestedScopes []string,
	ssoClientID string,
	accountID null.String,
) (bool, error) {
	// no legal scope requested = no scope mandatory to consent for the end-user
	reqLegalScopes := getLegalScopes(requestedScopes)
	if len(reqLegalScopes) == 0 {
		return true, nil
	}

	// no consent federation allowed for the soo client - no implicit consent
	// federation only enabled for the self client id
	if !(ssoClientID == afs.selfCliID) {
		return false, nil
	}
	// no linked account = no linked identities to make implicit the consent
	if !accountID.Valid {
		return false, nil
	}
	// on misakey client only, we auto-consent legal scopes considering linked identities
	// get consents for all identity linked to the account
	filters := domain.IdentityFilters{
		AccountID: accountID,
	}
	identities, err := afs.identityService.List(ctx, filters)
	if err != nil {
		return false, err
	}
	// retrieve consent session for all identities and check if a consent has already been done
	// for the requested legal scopes
	// NOTE: the following code does not handle the fact the end-user
	// has consented one scope on a specific client and one scope on another client.
	var legalOK bool
	for _, accountIdentity := range identities {
		// TODO: change on https://github.com/ory/hydra/issues/1926 release
		sessions, err := afs.authFlow.GetConsentSessions(ctx, accountIdentity.ID)
		if err != nil {
			return false, err
		}
		if legalOK = clientHasScopes(afs.selfCliID, sessions, reqLegalScopes); legalOK {
			break
		}
	}
	if !legalOK {
		return false, nil
	}
	// consider ourselves the consent can be skipped
	return true, nil
}

// ConsentRequiredErr helper
func (afs AuthFlowService) ConsentRequiredErr() string {
	return buildRedirectErr(merror.ConsentRequiredCode, "forbidden prompt=none", afs.consentPageURL)
}

// ConsentRedirectErr helper
func (afs AuthFlowService) ConsentRedirectErr(err error) string {
	return buildRedirectErr(merror.InvalidFlowCode, err.Error(), afs.consentPageURL)
}

// buildConsentURL
func (afs AuthFlowService) BuildConsentURL(consentChallenge string) string {
	// build the consent URL
	finalURL := *afs.consentPageURL

	// add consent_challenge to query params
	query := url.Values{}
	query.Set("consent_challenge", consentChallenge)

	// escape query parameters
	finalURL.RawQuery = query.Encode()
	return finalURL.String()

}

// AssertLegalScopes returns an error if any legal scopes contained in requested parameter
// is missing from the consented parameter
func AssertLegalScopes(requested []string, consented []string) error {
	requestedLegalScopes := getLegalScopes(requested)
	consentedLegalScopes := getLegalScopes(consented)
	if len(slice.StrIntersect(requestedLegalScopes, consentedLegalScopes)) != len(requestedLegalScopes) {
		return merror.Forbidden().
			Describe("some requested legal scopes have not been consented").
			Detail("requested_legal_scope", strings.Join(requestedLegalScopes, " ")).
			Detail("consented_legal_scope", strings.Join(consentedLegalScopes, " "))
	}
	return nil
}

func getLegalScopes(scopes []string) []string {
	legalScopes := []string{"tos", "privacy_policy"}
	return slice.StrIntersect(legalScopes, scopes)
}

func clientHasScopes(clientID string, sessions []consent.Session, scopes []string) bool {
	for _, session := range sessions {
		if session.ConsentRequest.Client.ID != clientID {
			continue
		}
		return len(slice.StrIntersect(scopes, session.GrantScope)) == len(scopes)
	}
	return false
}
