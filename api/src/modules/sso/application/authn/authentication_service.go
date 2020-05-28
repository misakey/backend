package authn

import (
	"context"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authentication"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type Service struct {
	steps        stepRepo
	codeValidity time.Duration
}

type stepRepo interface {
	Create(ctx context.Context, step *authn.Step) error
	CompleteAt(ctx context.Context, stepID int, completeTime time.Time) error
	Last(ctx context.Context, identityID string, methodName authn.Method) (authn.Step, error)
}

func NewService(steps stepRepo) Service {
	return Service{
		steps:        steps,
		codeValidity: 5 * time.Minute}
}

// AssertStep considering the method name and the received metadata
// Return no error in case of success
func (as *Service) AssertAuthnStep(ctx context.Context, assertion authn.Step) error {
	// always take the most recent step as the current one - ignore others
	currentStep, err := as.steps.Last(ctx, assertion.IdentityID, assertion.MethodName)
	if err != nil {
		return err
	}
	// check the most recent step has not been already complete
	if currentStep.Complete {
		return merror.Conflict().Describe("most recent step already complete")
	}

	// check the metadata
	var metadataErr error
	switch currentStep.MethodName {
	case authn.EmailedCodeMethod:
		metadataErr = as.assertEmailedCode(currentStep, assertion.Metadata)
	default:
		metadataErr = merror.BadRequest().Detail("method_name", merror.DVInvalid)
	}
	if metadataErr != nil {
		return metadataErr
	}

	// complete the authentication step - update the entity
	currentStep.Complete = true
	if err := as.steps.CompleteAt(ctx, currentStep.ID, time.Now()); err != nil {
		return err
	}

	return nil
}
