package authentication

import (
	"context"
	"fmt"
	"time"

	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/types"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authentication"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// CreateEmailedCode authentication step
func (as *Service) CreateEmailedCode(ctx context.Context, identityID string) error {
	// try to retrieve an existing code for this identity
	existing, err := as.steps.Last(ctx, identityID, authentication.EmailedCodeMethod)
	if err != nil && !merror.HasCode(err, merror.NotFoundCode) {
		return err
	}
	if err == nil && time.Now().Sub(existing.InitiatedAt) < as.codeValidity {
		return merror.Conflict().
			Describe("a code has already been generated").
			Detail("identity_id", merror.DVConflict).
			Detail("method_name", merror.DVConflict)
	}

	codeMetadata, err := generateCodeMetadata()
	if err != nil {
		return err
	}

	flow := authentication.Step{
		IdentityID: identityID,
		MethodName: authentication.EmailedCodeMethod,
		Metadata:   codeMetadata,

		CreatedAt: time.Now(),

		Complete:   false,
		CompleteAt: null.Time{},
	}
	if err := as.steps.Create(ctx, &flow); err != nil {
		return err
	}

	// TODO11: send the email
	fmt.Println("Emailed Code: ", string(codeMetadata))
	return nil
}

func (as *Service) assertEmailedCode(currentStep authentication.Step, inputMetadata types.JSON) error {
	// transform metadata into code metadata structure
	input, err := toCodeMetadata(inputMetadata)
	if err != nil {
		return merror.Forbidden().From(merror.OriBody).
			Describe(err.Error()).Detail("metadata", merror.DVMalformed)
	}
	stored, err := toCodeMetadata(currentStep.Metadata)
	if err != nil {
		return merror.Forbidden().
			Describef("could not convert step %d as emailed code: %v", currentStep.ID, err.Error()).
			Detail("stored_code", merror.DVMalformed)
	}

	// compare codes
	if input.Code != stored.Code {
		return merror.Forbidden().From(merror.OriBody).Detail("code", merror.DVInvalid)
	}

	// check stored code is not expired
	if time.Now().After(currentStep.InitiatedAt.Add(as.codeValidity)) {
		return merror.Forbidden().From(merror.OriBody).Detail("code", merror.DVExpired)
	}

	return nil
}
