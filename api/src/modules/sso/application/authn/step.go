package authn

import (
	"context"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
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

func (as *Service) InitStep(ctx context.Context, identity identity.Identity, methodName oidc.MethodRef) error {
	switch methodName {
	case oidc.AMREmailedCode:
		return as.CreateEmailedCode(ctx, identity)
	case oidc.AMRPrehashedPassword:
		return as.AssertPasswordExistence(ctx, identity)
	default:
		return merror.BadRequest().Describe("unknown method name").Detail("method_name", merror.DVInvalid)
	}

}

// AssertStep considering the method name and the received metadata
// It takes a pointer on the identity since the identity might be atlered by the authn step
// Return a nil error in case of success
func (as *Service) AssertStep(ctx context.Context, challenge string, identity *identity.Identity, assertion Step) error {
	// check the metadata
	var metadataErr error
	switch assertion.MethodName {
	case oidc.AMREmailedCode:
		metadataErr = as.assertEmailedCode(ctx, assertion)
	case oidc.AMRPrehashedPassword:
		metadataErr = as.assertPassword(ctx, *identity, assertion)
	case oidc.AMRAccountCreation:
		metadataErr = as.assertAccountCreation(ctx, challenge, identity, assertion)
	default:
		metadataErr = merror.BadRequest().Detail("method_name", merror.DVMalformed)
	}
	return metadataErr
}

// NextStep returns an Step according to ACR expectation and identity state
// without expectation:
// - return the preferredStep
// with expectations:
// - ACR1:
//     * return emailed_code
// - ACR2:
//     * without account:
//       -> unauthorized: return emailed_code
//       -> authorized: return account_creation
//     * with account: return prehashed_password
func (as *Service) NextStep(ctx context.Context, identity identity.Identity, currentACR oidc.ClassRef, expectations oidc.ClassRefs) (Step, error) {
	var err error
	var step Step

	switch expectations.Get() {
	case oidc.ACR1:
		err = as.prepareEmailedCode(ctx, identity, &step)
	case oidc.ACR2:
		// no linked account ? require one
		if identity.AccountID.IsZero() {
			err = as.requireAccount(ctx, identity, currentACR, &step)
		} else {
			err = as.preparePassword(ctx, identity, &step)
		}
	default:
		err = as.preferredStep(ctx, identity, &step)
	}
	step.IdentityID = identity.ID
	return step, err
}

func (as *Service) requireAccount(ctx context.Context, identity identity.Identity, currentACR oidc.ClassRef, step *Step) error {
	// if the ACR brought by authorization is less than 1, return an emailed code step to upgrade it
	if currentACR.LessThan(oidc.ACR1) {
		return as.prepareEmailedCode(ctx, identity, step)
	}

	// otherwise, ask for account creation
	step.MethodName = oidc.AMRAccountCreation
	step.RawJSONMetadata = nil
	return nil
}

// preferredStep is defined according to the identity state
// - has no account: emailed_code
// - has a linked account: prehashed_password
func (as *Service) preferredStep(ctx context.Context, identity identity.Identity, step *Step) error {
	// if the identity has no linked account, we automatically init a emailed code authentication step
	if identity.AccountID.IsZero() {
		return as.prepareEmailedCode(ctx, identity, step)
	}
	return as.preparePassword(ctx, identity, step)
}

func (as *Service) ExpireAll(ctx context.Context, identityID string) error {
	return as.steps.DeleteIncomplete(ctx, identityID)
}
