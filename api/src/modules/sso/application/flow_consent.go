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
		ID      string      `json:"client_id"`
		Name    string      `json:"name"`
		LogoURL null.String `json:"logo_uri"`
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
		v.Field(&cmd.ConsentedScopes, v.Required, v.Each(v.In("tos", "privacy_policy"))),
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

	// ensure tos and privacy policy have been accepted
	// NOTE: we might make optional these scopes for external RPs
	// it is currently mandatory for all of them
	var tosOK, ppOK bool
	for _, consentScope := range cmd.ConsentedScopes {
		tosOK = (tosOK || (consentScope == "tos"))
		ppOK = (ppOK || (consentScope == "privacy_policy"))
	}
	if !tosOK || !ppOK {
		return redirect, merror.Forbidden().
			Describe("tos and privacy_policy scopes must be consented").
			Detail("tos", merror.DVRequired).
			Detail("privacy_policy", merror.DVRequired)
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

	scopes := consentCtx.RequestedScope
	scopes = append(scopes, []string{"tos", "privacy_policy"}...)
	// 3. Check tos and privacy_policy consents
	if identity.AccountID.IsZero() {
		// get consents for the identity
		sessions, err := sso.authFlowService.ConsentGetSessions(ctx, identity.ID)
		if err != nil {
			return sso.authFlowService.ConsentRedirectErr(err)
		}
		tosOK, privacyPolicyOK := checkLegalScopes(consentCtx.Client.ID, sessions)
		if !tosOK || !privacyPolicyOK {
			return sso.authFlowService.BuildConsentURL(consentCtx.Challenge)
		}
	} else {
		// get consents for all identity linked to the account
		filters := domain.IdentityFilters{
			AccountID: identity.AccountID,
		}
		identities, err := sso.identityService.List(ctx, filters)
		if err != nil {
			return sso.authFlowService.ConsentRedirectErr(err)
		}
		var tosOK, privacyPolicyOK bool
		for _, accountIdentity := range identities {
			sessions, err := sso.authFlowService.ConsentGetSessions(ctx, accountIdentity.ID)
			if err != nil {
				return sso.authFlowService.ConsentRedirectErr(err)
			}
			idTosOK, idPrivacyPolicyOK := checkLegalScopes(consentCtx.Client.ID, sessions)
			tosOK = idTosOK || tosOK
			privacyPolicyOK = idPrivacyPolicyOK || privacyPolicyOK
		}

		if !tosOK || !privacyPolicyOK {
			return sso.authFlowService.BuildConsentURL(consentCtx.Challenge)
		}
	}

	// 4. build consent accept request directly
	acceptance := consent.Acceptance{
		GrantScope:  scopes, // accept all requested scopes
		Remember:    true,
		RememberFor: 0, // remember for ever the user consent
	}
	acceptance.Session.IDTokenClaims.Scope = strings.Join(scopes, " ")
	acceptance.Session.IDTokenClaims.Email = identifier.Value
	acceptance.Session.IDTokenClaims.AMR = consentCtx.AuthnContext.GetAMR()
	acceptance.Session.AccessTokenClaims.ACR = consentCtx.ACR

	// 5. tell hydra the consent contract
	return sso.authFlowService.ConsentAccept(ctx, consentChallenge, acceptance)
}

// checkLegalScopes
func checkLegalScopes(clientID string, sessions []consent.Session) (tosOK, privacyPolicyOK bool) {
	for _, session := range sessions {
		if session.ConsentRequest.Client.ID != clientID {
			continue
		}
		for _, scope := range session.GrantScope {
			if scope == "tos" {
				tosOK = true
				if privacyPolicyOK {
					// we don’t need to continue
					// if all the scopes are here
					return
				}
			}
			if scope == "privacy_policy" {
				privacyPolicyOK = true
				if tosOK {
					// we don’t need to continue
					// if all the scopes are here
					return
				}
			}
		}
	}
	return
}
