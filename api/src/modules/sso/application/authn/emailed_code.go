package authn

import (
	"context"
	"fmt"
	"time"

	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn/code"
)

// CreateEmailedCode authentication step
func (as *Service) CreateEmailedCode(ctx context.Context, identityID string) error {
	// try to retrieve an existing code for this identity
	existing, err := as.steps.Last(ctx, identityID, authn.EmailedCodeMethod)
	if err != nil && !merror.HasCode(err, merror.NotFoundCode) {
		return err
	}
	// if the last authn step is not complete and not expired, we can't create a new one
	if err == nil && !existing.Complete &&
		time.Now().Sub(existing.CreatedAt) < as.codeValidity {
		return merror.Conflict().
			Describe("a code has already been generated and not used").
			Detail("identity_id", merror.DVConflict).
			Detail("method_name", merror.DVConflict)
	}

	codeRawJSON, err := code.GenerateAsRawJSON()
	if err != nil {
		return err
	}

	flow := authn.Step{
		IdentityID:      identityID,
		MethodName:      authn.EmailedCodeMethod,
		RawJSONMetadata: codeRawJSON,

		CreatedAt: time.Now(),

		Complete:   false,
		CompleteAt: null.Time{},
	}
	if err := as.steps.Create(ctx, &flow); err != nil {
		return err
	}

	// retrieve the identifier
	identity, err := as.identityService.Get(ctx, identityID)
	if err != nil {
		// if the step exist, the identity should exist
		// if it does not, we have an internal problem
		return merror.Transform(err).Describe("could not find identity")
	}

	identifier, err := as.identifierService.Get(ctx, identity.IdentifierID)
	if err != nil {
		// here the identifier should exist
		// if it does not, we have an internal problem
		return merror.Transform(err).Describe("could not find identifier")
	}

	decodedCode, err := code.ToMetadata(codeRawJSON)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"to":   identifier.Value,
		"code": decodedCode.Code,
	}
	subject := fmt.Sprintf("Votre code de confirmation - %s", decodedCode.Code)
	content, err := as.templates.NewEmail(ctx, identifier.Value, subject, "code", data)
	if err != nil {
		return err
	}

	return as.emails.Send(ctx, content)
}

func (as *Service) assertEmailedCode(currentStep authn.Step, inputRawJSON types.JSON) error {
	// transform metadata into code metadata structure
	input, err := code.ToMetadata(inputRawJSON)
	if err != nil {
		return merror.Forbidden().From(merror.OriBody).
			Describe(err.Error()).Detail("metadata", merror.DVMalformed)
	}
	stored, err := code.ToMetadata(currentStep.RawJSONMetadata)
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
	if time.Now().After(currentStep.CreatedAt.Add(as.codeValidity)) {
		return merror.Forbidden().From(merror.OriBody).Detail("code", merror.DVExpired)
	}

	return nil
}
