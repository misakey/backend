package application

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/domain/consent"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// ConsentInitCmd ...
type ConsentInitCmd struct {
	ConsentChallenge string `query:"consent_challenge"`
}

// BindAndValidate ...
func (cmd *ConsentInitCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriQuery).Describe(err.Error())
	}

	return v.ValidateStruct(cmd,
		v.Field(&cmd.ConsentChallenge, v.Required),
	)
}

// InitConsent stage for a user (a.k.a. consent flow)
// It interacts with hydra to know either user has already consented to share data with the RP
// It returns a URL user's agent should be redirected to
// Today, it accept directly the consent of the user with the email claim contained in the ID token
func (sso *SSOService) InitConsent(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*ConsentInitCmd)

	// 1. get consent context
	consentCtx, err := sso.authFlowService.GetConsentContext(ctx, cmd.ConsentChallenge)
	if err != nil {
		return sso.authFlowService.ConsentRedirectErr(err), nil
	}

	// 2. retrieve subject information to put it in ID tokens as claims
	curIdentity, err := identity.Get(ctx, sso.sqlDB, consentCtx.OIDCContext.MID())
	if err != nil {
		if merror.HasCode(err, merror.NotFoundCode) {
			return sso.authFlowService.BuildResetURL(consentCtx.RequestURL), nil
		}
		return sso.authFlowService.ConsentRedirectErr(err), nil
	}

	// 3. upsert the Sec Level authentication session
	// this is the first time we receive a potentially new login session id
	session := authn.Session{
		ID:          consentCtx.LoginSessionID,
		ACR:         consentCtx.ACR,
		RememberFor: consentCtx.ACR.RememberFor(),
		IdentityID:  consentCtx.OIDCContext.MID(),
		AccountID:   consentCtx.OIDCContext.AID(),
	}
	if err := sso.AuthenticationService.UpsertSession(ctx, session); err != nil {
		return sso.authFlowService.ConsentRedirectErr(err), nil
	}

	// 4. ask our consent service if the end-user manual consent can be skipped
	skip, err := sso.authFlowService.ShouldSkipConsent(
		ctx, sso.sqlDB,
		consentCtx.RequestedScope, consentCtx.Client.ID, curIdentity.AccountID,
	)
	if err != nil {
		return sso.authFlowService.ConsentRedirectErr(err), nil
	}

	// consider both our's and hydra's decision about skipping the manual consent
	if skip || consentCtx.Skip {
		return sso.authFlowService.BuildAndAcceptConsent(ctx, consentCtx, curIdentity.Identifier.Value), nil
	}

	if authflow.HasNonePrompt(consentCtx.RequestURL) {
		return sso.authFlowService.ConsentRequiredErr(), nil
	}

	return sso.authFlowService.BuildConsentURL(consentCtx.Challenge), nil
}

// ConsentInfoQuery ...
type ConsentInfoQuery struct {
	ConsentChallenge string `query:"consent_challenge"`
}

// BindAndValidate ...
func (query *ConsentInfoQuery) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(query); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	return v.ValidateStruct(query,
		v.Field(&query.ConsentChallenge, v.Required),
	)
}

// ConsentInfoView bears data about current user authentication status
type ConsentInfoView struct {
	Subject        string        `json:"subject"`
	ACR            oidc.ClassRef `json:"acr"`
	RequestedScope []string      `json:"scope"`
	OIDCContext    oidc.Context  `json:"context"`
	Client         struct {
		ID        string      `json:"id"`
		Name      string      `json:"name"`
		LogoURL   null.String `json:"logo_uri"`
		TosURL    null.String `json:"tos_uri"`
		PolicyURL null.String `json:"policy_uri"`
	} `json:"client"`
}

// GetConsentInfo ...
func (sso *SSOService) GetConsentInfo(ctx context.Context, gen request.Request) (interface{}, error) {
	query := gen.(*ConsentInfoQuery)

	view := ConsentInfoView{}

	consentCtx, err := sso.authFlowService.GetConsentContext(ctx, query.ConsentChallenge)
	if err != nil {
		return view, merror.Transform(err).Describe("could not get context")
	}

	// fill view with domain model
	view.Subject = consentCtx.Subject
	view.ACR = consentCtx.ACR
	view.RequestedScope = consentCtx.RequestedScope
	view.OIDCContext = consentCtx.OIDCContext
	view.Client.ID = consentCtx.Client.ID
	view.Client.Name = consentCtx.Client.Name
	view.Client.LogoURL = consentCtx.Client.LogoURL
	view.Client.TosURL = consentCtx.Client.TosURL
	view.Client.PolicyURL = consentCtx.Client.PolicyURL
	return view, nil
}

// ConsentAcceptCmd ...
type ConsentAcceptCmd struct {
	IdentityID       string   `json:"identity_id"`
	ConsentChallenge string   `json:"consent_challenge"`
	ConsentedScopes  []string `json:"consented_scopes"`
}

// BindAndValidate ...
func (cmd *ConsentAcceptCmd) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(cmd); err != nil {
		return merror.BadRequest().From(merror.OriBody).Describe(err.Error())
	}

	return v.ValidateStruct(cmd,
		v.Field(&cmd.IdentityID, v.Required, is.UUIDv4),
		v.Field(&cmd.ConsentChallenge, v.Required),
	)
}

// AcceptConsent ...
func (sso *SSOService) AcceptConsent(ctx context.Context, gen request.Request) (interface{}, error) {
	cmd := gen.(*ConsentAcceptCmd)
	redirect := consent.Redirect{}
	// 1. get consent context
	consentCtx, err := sso.authFlowService.GetConsentContext(ctx, cmd.ConsentChallenge)
	if err != nil {
		return redirect, err
	}

	// 2. retrieve subject information to put it in ID tokens as claims
	curIdentity, err := identity.Get(ctx, sso.sqlDB, consentCtx.OIDCContext.MID())
	if err != nil {
		return redirect, err
	}
	if curIdentity.ID != cmd.IdentityID {
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
	redirect.To = sso.authFlowService.BuildAndAcceptConsent(ctx, consentCtx, curIdentity.Identifier.Value)
	return redirect, nil
}
