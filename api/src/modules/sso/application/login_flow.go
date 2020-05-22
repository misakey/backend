package application

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authentication"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

func (sso SSOService) LoginInit(ctx context.Context, loginChallenge string) string {
	return sso.authFlowService.LoginInit(ctx, loginChallenge)
}

type LoginStepCmd struct {
	LoginChallenge string              `json:"login_challenge"`
	Step           authentication.Step `json:"step"`
}

// Validate the LoginStepCmd
func (cmd LoginStepCmd) Validate() error {
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

// LoginStep assert an authentication step in a multi-factor authentication process
// Today there is only one-step authentication process existing
func (sso SSOService) LoginStep(ctx context.Context, cmd LoginStepCmd) (login.Redirect, error) {
	redirect := login.Redirect{}

	// 1. ensure the login challenge is correct
	logCtx, err := sso.authFlowService.LoginGetContext(ctx, cmd.LoginChallenge)
	if err != nil {
		return redirect, err
	}

	// 2. try to assert the authentication step
	if err := sso.authenticationService.AssertStep(ctx, cmd.Step); err != nil {
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
