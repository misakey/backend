package authn

import (
	"context"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/adaptor/email"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/account"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type Service struct {
	steps             stepRepo
	identifierService identifier.IdentifierService
	identityService   identity.IdentityService
	accountService    account.AccountService
	templates         email.Renderer
	emails            email.Sender
	codeValidity      time.Duration
}

type stepRepo interface {
	Create(ctx context.Context, step *authn.Step) error
	CompleteAt(ctx context.Context, stepID int, completeTime time.Time) error
	Last(ctx context.Context, identityID string, methodName authn.MethodRef) (authn.Step, error)
}

func NewService(
	steps stepRepo,
	identifierService identifier.IdentifierService,
	identityService identity.IdentityService,
	accountService account.AccountService,
	templates email.Renderer,
	emails email.Sender) Service {
	return Service{
		steps:             steps,
		identifierService: identifierService,
		identityService:   identityService,
		accountService:    accountService,
		templates:         templates,
		emails:            emails,
		codeValidity:      5 * time.Minute,
	}
}

// AssertStep considering the method name and the received metadata
// Return no error in case of success
func (as *Service) AssertAuthnStep(ctx context.Context, assertion authn.Step) (authn.ClassRef, authn.MethodRefs, error) {
	acr := authn.ACR0
	amr := authn.MethodRefs{}

	// check the metadata
	var metadataErr error
	switch assertion.MethodName {
	case authn.AMREmailedCode:
		metadataErr = as.assertEmailedCode(ctx, assertion)
		acr = authn.ACR1
	case authn.AMRPrehashedPassword:
		metadataErr = as.assertPassword(ctx, assertion)
		acr = authn.ACR2
	default:
		metadataErr = merror.BadRequest().Detail("method_name", merror.DVMalformed)
	}
	amr.Add(assertion.MethodName)
	return acr, amr, metadataErr
}

// GetRememberFor as an integer corresponding to seconds, according to the authentication context class
func (as *Service) GetRememberFor(acr authn.ClassRef) int {
	switch acr {
	case authn.ACR1:
		return 3600 // 1h
	case authn.ACR2:
		return 2592000 // 30d
	}
	return 1
}
