package authn

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// Step in a multi-factor authentication process.
type Step struct {
	ID              int
	IdentityID      string         `json:"identity_id"`
	MethodName      oidc.MethodRef `json:"method_name"`
	RawJSONMetadata types.JSON     `json:"metadata"`
	CreatedAt       time.Time
	Complete        bool
	CompleteAt      null.Time
}

// InitStep ...
func (as *Service) InitStep(
	ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client,
	identity identity.Identity, methodName oidc.MethodRef,
) error {
	switch methodName {
	case oidc.AMREmailedCode:
		_, err := prepareEmailedCode(ctx, as, exec, redConn, identity, oidc.ACR0, &Step{}, false)
		return err
	case oidc.AMRPrehashedPassword:
		return assertPasswordExistence(ctx, identity)
	case oidc.AMRWebauthn:
		return assertWebauthnCredentials(ctx, exec, identity)
	case oidc.AMRTOTP:
		return assertTOTPSecret(ctx, exec, identity)
	default:
		return merr.BadRequest().Desc("cannot init method").Add("method_name", merr.DVInvalid)
	}
}

// AssertStep considering the method name and the received metadata
// It takes a pointer on the identity since the identity might be atlered by the authn step
// Return a nil error in case of success
func (as *Service) AssertStep(
	ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client,
	challenge string, identity *identity.Identity, assertion Step,
) error {
	// check the metadata
	var metadataErr error
	switch assertion.MethodName {
	case oidc.AMREmailedCode:
		metadataErr = as.assertEmailedCode(ctx, exec, assertion)
	case oidc.AMRPrehashedPassword:
		metadataErr = as.assertPassword(ctx, exec, *identity, assertion)
	case oidc.AMRAccountCreation:
		metadataErr = as.assertAccountCreation(ctx, exec, redConn, challenge, identity, assertion)
	case oidc.AMRWebauthn:
		metadataErr = as.assertWebauthn(ctx, exec, redConn, *identity, assertion)
	case oidc.AMRTOTP:
		metadataErr = as.assertTOTP(ctx, exec, redConn, *identity, assertion)
	case oidc.AMRResetPassword:
		metadataErr = as.resetPassword(ctx, exec, redConn, challenge, *identity, assertion)
	default:
		metadataErr = merr.BadRequest().Add("method_name", merr.DVMalformed)
	}
	return metadataErr
}

type authnMethodHandler func(
	context.Context, *Service, boil.ContextExecutor, *redis.Client,
	identity.Identity, oidc.ClassRef, *Step,
	bool,
) (*Step, error)

var prepareStepFunc = map[oidc.MethodRef]authnMethodHandler{
	oidc.AMREmailedCode:       prepareEmailedCode,
	oidc.AMRPrehashedPassword: preparePassword,
	oidc.AMRTOTP:              prepareTOTP,
	oidc.AMRWebauthn:          prepareWebauthn,
}

// PrepareNextStep returns a prepared authn Step according to
// the current ACR, expected ACR, and the identity state
// it return a nil step when no step are required anymore
//
// see https://backend.docs.misakey.dev/concepts/authorization-and-authentication/#43-methods for more details about ruling
func (as *Service) PrepareNextStep(
	ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client,
	identity identity.Identity, currentACR oidc.ClassRef, expectedACR oidc.ClassRef,
	passwordReset bool,
) (*Step, error) {
	var step Step

	nextMethod := oidc.GetNextMethod(currentACR, expectedACR)
	if nextMethod == nil {
		return nil, nil
	}

	step.MethodName = *nextMethod
	step.IdentityID = identity.ID
	return prepareStepFunc[step.MethodName](ctx, as, exec, redConn, identity, currentACR, &step, passwordReset)
}

// ExpireAll ...
func (as *Service) ExpireAll(ctx context.Context, exec boil.ContextExecutor, identityID string) error {
	return deleteIncompleteSteps(ctx, exec, identityID)
}
