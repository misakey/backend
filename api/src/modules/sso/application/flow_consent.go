package application

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/consent"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/oidc"
)

// ConsentInfoView bears data about current user authentication status
type ConsentInfoView struct {
	Subject        string        `json:"subject"`
	ACR            oidc.ClassRef `json:"acr"`
	RequestedScope []string      `json:"scope"`
	AuthnContext   oidc.Context  `json:"context"`
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

	consentCtx, err := sso.authFlowService.GetConsentContext(ctx, loginChallenge)
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

// Init a user consent stage (a.k.a. consent flow)
// It interacts with hydra to know either user has already consented to share data with the RP
// It returns a URL user's agent should be redirected to
// Today, it accept directly the consent of the user with the email claim contained in the ID token
func (sso SSOService) ConsentInit(ctx context.Context, consentChallenge string) string {
	// 1. get consent context
	consentCtx, err := sso.authFlowService.GetConsentContext(ctx, consentChallenge)
	if err != nil {
		return sso.authFlowService.ConsentRedirectErr(err)
	}

	// 2. retrieve subject information to put it in ID tokens as claims
	identity, err := sso.identityService.Get(ctx, consentCtx.Subject)
	if err != nil {
		return sso.authFlowService.ConsentRedirectErr(err)
	}

	// 3. upsert the Sec Level authentication session
	// this is the first time we receive a potentially new login session id
	session := authn.Session{
		ID:          consentCtx.LoginSessionID,
		ACR:         consentCtx.ACR,
		RememberFor: consentCtx.ACR.RememberFor(),
	}
	if err := sso.authenticationService.UpsertSession(ctx, session); err != nil {
		return sso.authFlowService.ConsentRedirectErr(err)
	}

	// 4. ask our consent service if the end-user manual consent can be skipped
	skip, err := sso.authFlowService.ShouldSkipConsent(
		ctx, consentCtx.RequestedScope,
		consentCtx.Client.ID,
		identity.AccountID,
	)
	if err != nil {
		return sso.authFlowService.ConsentRedirectErr(err)
	}

	// consider both our's and hydra's decision about skipping the manual consent
	if skip || consentCtx.Skip {
		return sso.authFlowService.BuildAndAcceptConsent(ctx, consentCtx, identity.Identifier.Value)
	}

	if authflow.NonePrompt(consentCtx.RequestURL) {
		return sso.authFlowService.ConsentRequiredErr()
	}

	return sso.authFlowService.BuildConsentURL(consentCtx.Challenge)
}

func (sso SSOService) ConsentAccept(ctx context.Context, cmd ConsentAcceptCmd) (consent.Redirect, error) {
	redirect := consent.Redirect{}
	// 1. get consent context
	consentCtx, err := sso.authFlowService.GetConsentContext(ctx, cmd.ConsentChallenge)
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
	if err := authflow.AssertLegalScopes(consentCtx.RequestedScope, cmd.ConsentedScopes); err != nil {
		return redirect, err
	}

	// override requested scope with the final built consented scopes
	consentCtx.RequestedScope = append(consentScopes, cmd.ConsentedScopes...)

	// 4. tell hydra the consent contract & returns the hydra url response
	redirect.To = sso.authFlowService.BuildAndAcceptConsent(ctx, consentCtx, identity.Identifier.Value)
	return redirect, nil
}
