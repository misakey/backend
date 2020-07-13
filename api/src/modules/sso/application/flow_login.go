package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

// Init a user authentication stage (a.k.a. login flow)
// It interacts with hydra and login sessions to know either user is already authenticated or not
// It returns a URL user's agent should be redirected to
func (sso SSOService) LoginInit(ctx context.Context, loginChallenge string) string {
	// get info about current login session
	loginCtx, err := sso.authFlowService.GetLoginContext(ctx, loginChallenge)
	if err != nil {
		return sso.authFlowService.LoginRedirectErr(err)
	}

	// skip indicates if an active session has been detected
	// we check if login session ACR are high enough to accept authentication
	if loginCtx.Skip {
		session, err := sso.authenticationService.GetSession(ctx, loginCtx.SessionID)
		if err == nil {
			// if the session ACR is higher or equivalent to the expected ACR, we accept the login
			if session.ACR >= loginCtx.OIDCContext.ACRValues.Get() {
				// set browser cookie as authentication method
				// TODO add it to login flow
				loginCtx.OIDCContext.AMRs.Add(authn.AMRBrowserCookie)
				loginCtx.OIDCContext.ACRValues.Set(session.ACR)
				redirect, err := sso.authFlowService.BuildAndAcceptLogin(ctx, loginCtx)
				if err != nil {
					return sso.authFlowService.LoginRedirectErr(err)
				}
				return redirect.To
			}
		}
		if authflow.NonePrompt(loginCtx.RequestURL) {
			return sso.authFlowService.LoginRequiredErr()
		}
	}
	return sso.authFlowService.BuildLoginURL(loginChallenge)
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
	RequestedScope []string        `json:"scope"`
	ACRValues      authn.ClassRefs `json:"acr_values"`
	LoginHint      string          `json:"login_hint"`
}

func (sso SSOService) LoginInfo(ctx context.Context, loginChallenge string) (LoginInfoView, error) {
	view := LoginInfoView{}

	logCtx, err := sso.authFlowService.GetLoginContext(ctx, loginChallenge)
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
	logCtx, err := sso.authFlowService.GetLoginContext(ctx, cmd.LoginChallenge)
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
	acr, err := sso.authenticationService.AssertAuthnStep(ctx, cmd.Step)
	if err != nil {
		return redirect, err
	}

	// 3. if the authn step is emailed_code - we confirm the identity
	if cmd.Step.MethodName == authn.AMREmailedCode {
		// handle the reset password extension only if the emailed_code method has been used
		if cmd.PasswordResetExt != nil {
			if err := sso.resetPassword(ctx, *cmd.PasswordResetExt, cmd.Step.IdentityID); err != nil {
				return redirect, err
			}
			acr = authn.ACR2
		}
	}

	// 4. accept the login session
	logCtx.Subject = cmd.Step.IdentityID
	logCtx.OIDCContext.ACRValues.Set(acr)
	logCtx.OIDCContext.AMRs.Add(cmd.Step.MethodName)
	return sso.authFlowService.BuildAndAcceptLogin(ctx, logCtx)
}
