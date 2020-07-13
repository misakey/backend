package application

import (
	"context"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/consent"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

// ConsentInfoView bears data about current user authentication status
type ConsentInfoView struct {
	Subject        string        `json:"subject"`
	ACR            string        `json:"acr"`
	RequestedScope []string      `json:"scope"`
	AuthnContext   authn.Context `json:"context"`
	Client         struct {
		ID        string      `json:"id"`
		Name      string      `json:"name"`
		LogoURL   null.String `json:"logo_uri"`
		TosURL    null.String `json:"tos_uri"`
		PolicyURL null.String `json:"policy_uri"`
	} `json:"client"`
}

func (sso SSOService) ConsentInfo(ctx context.Context, loginChallenge string) (ConsentInfoView, error) {
	view := ConsentInfoView{}

	consentCtx, err := sso.authFlowService.ConsentGetContext(ctx, loginChallenge)
	if err != nil {
		return view, merror.Transform(err).Describe("could not get context")
	}

	// fill view with domain model
	view.Subject = consentCtx.Subject
	view.ACR = consentCtx.ACR
	view.RequestedScope = consentCtx.RequestedScope
	view.AuthnContext = consentCtx.AuthnContext
	view.Client.ID = consentCtx.Client.ID
	view.Client.Name = consentCtx.Client.Name
	view.Client.LogoURL = consentCtx.Client.LogoURL
	view.Client.TosURL = consentCtx.Client.TosURL
	view.Client.PolicyURL = consentCtx.Client.PolicyURL
	return view, nil
}

type ConsentAcceptCmd struct {
	IdentityID       string   `json:"identity_id"`
	ConsentChallenge string   `json:"consent_challenge"`
	ConsentedScopes  []string `json:"consented_scopes"`
}

func (cmd ConsentAcceptCmd) Validate() error {

	return v.ValidateStruct(&cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4.Error("identity id should be an uuid v4")),
		v.Field(&cmd.ConsentChallenge, v.Required),
	)
}

func (sso SSOService) ConsentAccept(ctx context.Context, cmd ConsentAcceptCmd) (login.Redirect, error) {
	redirect := login.Redirect{}
	// 1. get consent context
	consentCtx, err := sso.authFlowService.ConsentGetContext(ctx, cmd.ConsentChallenge)
	if err != nil {
		return redirect, err
	}

	// 2. retrieve subject information to put it in ID tokens as claims
	identity, err := sso.identityService.Get(ctx, consentCtx.Subject)
	if err != nil {
		return redirect, err
	}
	if identity.ID != cmd.IdentityID {
		return redirect, merror.Forbidden().Detail("identity_id", merror.DVForbidden)
	}
	identifier, err := sso.identifierService.Get(ctx, identity.IdentifierID)
	if err != nil {
		return redirect, err
	}

	// 3. build consent accept
	var consentScopes []string
	for _, reqScope := range consentCtx.RequestedScope {
		// automatically add openid scope
		// because there is no need to consent to this one
		if reqScope == "openid" {
			consentScopes = append(consentScopes, reqScope)
		}
	}

	// ensure requested legal scopes have been consented
	if err := assertLegalScopes(consentCtx.RequestedScope, cmd.ConsentedScopes); err != nil {
		return redirect, err
	}

	consentScopes = append(consentScopes, cmd.ConsentedScopes...)
	acceptance := consent.Acceptance{
		GrantScope:  consentScopes,
		Remember:    true,
		RememberFor: 0, // remember for ever the user consent
	}
	acceptance.Session.IDTokenClaims.Scope = strings.Join(consentScopes, " ")
	acceptance.Session.IDTokenClaims.Email = identifier.Value
	acceptance.Session.IDTokenClaims.AMR = consentCtx.AuthnContext.GetAMR()
	acceptance.Session.AccessTokenClaims.ACR = consentCtx.ACR

	// 4. tell hydra the consent contract
	redirect.To = sso.authFlowService.ConsentAccept(ctx, cmd.ConsentChallenge, acceptance)
	return redirect, nil
}

// Init a user consent stage (a.k.a. consent flow)
// It interacts with hydra to know either user has already consented to share data with the RP
// It returns a URL user's agent should be redirected to
// Today, it accept directly the consent of the user with the email claim contained in the ID token
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

	// 3. handle misakey auto-consent cross-identity for the same client id
	reqLegalScopes, hasLegScope := getLegalScopes(consentCtx.RequestedScope)
	isMisakey := (consentCtx.Client.ID == sso.selfCliID)
	if isMisakey && hasLegScope && identity.AccountID.Valid && !consentCtx.Skip {
		// on misakey client only, we auto-consent legal scopes considering linked identities
		// get consents for all identity linked to the account
		filters := domain.IdentityFilters{
			AccountID: identity.AccountID,
		}
		identities, err := sso.identityService.List(ctx, filters)
		if err != nil {
			return sso.authFlowService.ConsentRedirectErr(err)
		}
		// retrieve consent session for all identities and check if a consent has already been done
		// for the requested legal scopes
		// NOTE: the following code does not handle the fact the end-user
		// has consented one scope on a specific client and one scope on another client.
		var legalOK bool
		for _, accountIdentity := range identities {
			sessions, err := sso.authFlowService.ConsentGetSessions(ctx, accountIdentity.ID)
			if err != nil {
				return sso.authFlowService.ConsentRedirectErr(err)
			}
			if legalOK = clientHasScopes(sso.selfCliID, sessions, reqLegalScopes); legalOK {
				break
			}
		}
		if !legalOK {
			return sso.authFlowService.BuildConsentURL(consentCtx.Challenge)
		}
		// consider ourselves the consent can be skipped
		consentCtx.Skip = true
	}

	// if consent is not skipped, we redirect to the consent page
	if !consentCtx.Skip {
		return sso.authFlowService.BuildConsentURL(consentCtx.Challenge)
	}

	// 4. build consent accept request directly
	acceptance := consent.Acceptance{
		GrantScope:  consentCtx.RequestedScope, // accept all requested scopes
		Remember:    true,
		RememberFor: 0, // remember for ever the user consent
	}
	acceptance.Session.IDTokenClaims.Scope = strings.Join(consentCtx.RequestedScope, " ")
	acceptance.Session.IDTokenClaims.Email = identifier.Value
	acceptance.Session.IDTokenClaims.AMR = consentCtx.AuthnContext.GetAMR()
	acceptance.Session.AccessTokenClaims.ACR = consentCtx.ACR

	// 5. tell hydra the consent contract
	return sso.authFlowService.ConsentAccept(ctx, consentChallenge, acceptance)
}

func assertLegalScopes(requested []string, consented []string) error {
	requestedLegalScopes, _ := getLegalScopes(requested)
	consentedLegalScopes, _ := getLegalScopes(consented)
	if len(intersect(requestedLegalScopes, consentedLegalScopes)) != len(requestedLegalScopes) {
		return merror.Forbidden().
			Describe("some requested legal scopes have not been consented").
			Detail("requested_legal_scope", strings.Join(requestedLegalScopes, " ")).
			Detail("consented_legal_scope", strings.Join(consentedLegalScopes, " "))
	}
	return nil
}

func getLegalScopes(scopes []string) ([]string, bool) {
	legalScopes := []string{"tos", "privacy_policy"}
	inter := intersect(legalScopes, scopes)
	return inter, len(inter) > 0
}

func clientHasScopes(clientID string, sessions []consent.Session, scopes []string) bool {
	for _, session := range sessions {
		if session.ConsentRequest.Client.ID != clientID {
			continue
		}
		return len(intersect(scopes, session.GrantScope)) == len(scopes)
	}
	return false
}

func intersect(a []string, b []string) []string {
	var inter []string
	for i := 0; i < len(a); i++ {
		for y := 0; y < len(b); y++ {
			if b[y] == a[i] {
				inter = append(inter, a[i])
				break
			}
		}
	}
	return inter
}
