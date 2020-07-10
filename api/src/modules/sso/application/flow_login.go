package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

func (sso SSOService) LoginInit(ctx context.Context, loginChallenge string) string {
	return sso.authFlowService.LoginInit(ctx, loginChallenge)
}

// LoginInfoView bears data about current user authentication status
type LoginInfoView struct {
	Client struct { // concerned relying party
		ID        string      `json:"id"`
		Name      string      `json:"name"`
		LogoURL   null.String `json:"logo_uri"`
		TosURL    null.String `json:"tos_uri"`
		PolicyURL null.String `json:"policy_uri"`
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
	view.Client.TosURL = logCtx.Client.TosURL
	view.Client.PolicyURL = logCtx.Client.PolicyURL
	view.RequestedScope = logCtx.RequestedScope
	view.ACRValues = logCtx.OIDCContext.ACRValues
	view.LoginHint = logCtx.OIDCContext.LoginHint
	return view, nil
}

type LoginAuthnStepCmd struct {
	LoginChallenge   string            `json:"login_challenge"`
	Step             authn.Step        `json:"authn_step"`
	PasswordResetExt *PasswordResetCmd `json:"password_reset"`
}

// Validate the LoginStepCmd
func (cmd LoginAuthnStepCmd) Validate() error {
	// validate nested structure separately
	if err := v.ValidateStruct(&cmd.Step,
		v.Field(&cmd.Step.IdentityID, v.Required, is.UUIDv4.Error("identity id should be an uuid v4")),
		v.Field(&cmd.Step.MethodName, v.Required),
		v.Field(&cmd.Step.RawJSONMetadata, v.Required),
	); err != nil {
		return err
	}

	if err := v.ValidateStruct(&cmd,
		v.Field(&cmd.LoginChallenge, v.Required),
	); err != nil {
		return err
	}

	if cmd.PasswordResetExt != nil {
		return cmd.PasswordResetExt.Validate()
	}

	return nil
}

// LoginStep assert an authentication step in a multi-factor authentication process
// Today there is only one-step authentication process existing
func (sso SSOService) LoginAuthnStep(ctx context.Context, cmd LoginAuthnStepCmd) (login.Redirect, error) {
	redirect := login.Redirect{}

	// 1. ensure the login challenge is correct and the identity is authable
	logCtx, err := sso.authFlowService.LoginGetContext(ctx, cmd.LoginChallenge)
	if err != nil {
		return redirect, err
	}
	identity, err := sso.identityService.Get(ctx, cmd.Step.IdentityID)
	if err != nil {
		return redirect, err
	}
	if !identity.IsAuthable {
		return redirect, merror.Forbidden().Describe("identity not authable")
	}

	// 2. try to assert the authentication step
	acr, amr, err := sso.authenticationService.AssertAuthnStep(ctx, cmd.Step)
	if err != nil {
		return redirect, err
	}

	// 3. if the authn step was emailed_code - we confirm the identity
	if amr.Has(authn.AMREmailedCode) {
		if err := sso.identityService.Confirm(ctx, cmd.Step.IdentityID); err != nil {
			return redirect, err
		}

		// handle the reset password extension only if the emailed_code method has been used
		if cmd.PasswordResetExt != nil {
			if err := sso.resetPassword(ctx, *cmd.PasswordResetExt, cmd.Step.IdentityID); err != nil {
				return redirect, err
			}
			acr = authn.ACR2
		}
	}

	// 4. accept the login session
	acceptance := login.Acceptance{
		// TODO: handle session for identity ID corresponding to same accounts
		Subject: cmd.Step.IdentityID,

		Remember:    true,
		RememberFor: sso.authenticationService.GetRememberFor(acr),

		ACR:     acr.String(),
		Context: authn.NewContext().SetAMR(amr),
	}
	return sso.authFlowService.LoginAccept(ctx, logCtx.Challenge, acceptance)
}
