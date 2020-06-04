package application

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

func (sso SSOService) LoginInit(ctx context.Context, loginChallenge string) string {
	return sso.authFlowService.LoginInit(ctx, loginChallenge)
}

type LoginAuthnStepCmd struct {
	LoginChallenge string     `json:"login_challenge"`
	Step           authn.Step `json:"authn_step"`
}

// Validate the LoginStepCmd
func (cmd LoginAuthnStepCmd) Validate() error {
	// validate nested structure separately
	if err := validation.ValidateStruct(&cmd.Step,
		validation.Field(&cmd.Step.IdentityID, validation.Required, is.UUIDv4.Error("identity id should be an uuid v4")),
		validation.Field(&cmd.Step.MethodName, validation.Required),
		validation.Field(&cmd.Step.Metadata, validation.Required),
	); err != nil {
		return err
	}

	if err := validation.ValidateStruct(&cmd,
		validation.Field(&cmd.LoginChallenge, validation.Required),
	); err != nil {
		return err
	}

	return nil
}

// LoginInfoView bears data about current user authentication status
type LoginInfoView struct {
	Client struct { // concerned relying party
		ID      string      `json:"id"`
		Name    string      `json:"name"`
		LogoURL null.String `json:"logo_uri"`
	} `json:"client"`
	RequestedScope []string `json:"scope"`
	ACRValues      []string `json:"acr_values"`
	LoginHint      string   `json:"login_hint"`
}

func (sso SSOService) LoginInfo(ctx context.Context, loginChallenge string) (LoginInfoView, error) {
	view := LoginInfoView{}

	logCtx, err := sso.authFlowService.LoginGetContext(ctx, loginChallenge)
	if err != nil {
		return view, merror.Transform(err).Describe("could not get context")
	}

	// fill view with domain model
	view.Client.ID = logCtx.Client.ID
	view.Client.Name = logCtx.Client.Name
	view.Client.LogoURL = logCtx.Client.LogoURL
	view.RequestedScope = logCtx.RequestedScope
	view.ACRValues = logCtx.OIDCContext.ACRValues
	view.LoginHint = logCtx.OIDCContext.LoginHint
	return view, nil
}

// LoginStep assert an authentication step in a multi-factor authentication process
// Today there is only one-step authentication process existing
func (sso SSOService) LoginAuthnStep(ctx context.Context, cmd LoginAuthnStepCmd) (login.Redirect, error) {
	redirect := login.Redirect{}

	// 1. ensure the login challenge is correct
	logCtx, err := sso.authFlowService.LoginGetContext(ctx, cmd.LoginChallenge)
	if err != nil {
		return redirect, err
	}

	// 2. try to assert the authentication step
	if err := sso.authenticationService.AssertAuthnStep(ctx, cmd.Step); err != nil {
		return redirect, err
	}

	// 3. confirm the identity
	// TODO: just call it when it is necessary ?
	if err := sso.identityService.Confirm(ctx, cmd.Step.IdentityID); err != nil {
		return redirect, err
	}

	// 4. accept the login session
	acceptance := login.Acceptance{
		// TODO: handle session for identity ID corresponding to same accounts - there is no account today
		Subject: cmd.Step.IdentityID,
		// TODO: make authentication service evaluate the real ACR the day we introduce passwords
		ACR:         "1",
		Remember:    true,
		RememberFor: 2592000,
	}
	return sso.authFlowService.LoginAccept(ctx, logCtx.Challenge, acceptance)
}
