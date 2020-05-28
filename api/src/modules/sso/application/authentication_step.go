package application

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authentication"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// AuthenticationStepCmd orders:
// - the retry of an authentication step init for the identity
type AuthenticationStepCmd struct {
	LoginChallenge string              `json:"login_challenge"`
	Step           authentication.Step `json:"step"`
}

// Validate the AuthenticationStepCmd
func (cmd AuthenticationStepCmd) Validate() error {
	if err := validation.ValidateStruct(&cmd.Step,
		validation.Field(&cmd.Step.IdentityID, validation.Required, is.UUIDv4.Error("identity_id must be an UUIDv4")),
		validation.Field(&cmd.Step.MethodName, validation.Required),
	); err != nil {
		return err
	}

	return validation.ValidateStruct(&cmd,
		validation.Field(&cmd.LoginChallenge, validation.Required),
	)
}

// This method is used to try to init an authentication step
func (sso SSOService) InitStep(ctx context.Context, cmd AuthenticationStepCmd) error {
	var err error

	// 0. check if the identity exists
	_, err = sso.identityService.Get(ctx, cmd.Step.IdentityID)
	if err != nil {
		return err
	}

	// 1. check login challenge
	_, err = sso.authFlowService.LoginGetContext(ctx, cmd.LoginChallenge)
	if err != nil {
		return merror.NotFound().Describe("could not find login challenge").Detail("login_challenge", merror.DVNotFound)
	}

	// 2. we try to init the authentication step
	switch cmd.Step.MethodName {
	case authentication.EmailedCodeMethod:
		return sso.authenticationService.CreateEmailedCode(ctx, cmd.Step.IdentityID)
	default:
		return merror.BadRequest().Describe("unknown method name").Detail("method_name", merror.DVInvalid)
	}
}
