package authn

import (
	"context"
	"fmt"
	"time"

	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authn/code"
)

// CreateEmailedCode authentication step
func (as *Service) CreateEmailedCode(ctx context.Context, identityID string) error {
	// try to retrieve an existing code for this identity
	existing, err := as.steps.Last(ctx, identityID, authn.AMREmailedCode)
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
		MethodName:      authn.AMREmailedCode,
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
		return merror.Transform(err).Describe("get identity")
	}

	identifier, err := as.identifierService.Get(ctx, identity.IdentifierID)
	if err != nil {
		return merror.Transform(err).Describe("get identifier")
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

	if err := as.emails.Send(ctx, content); err != nil {
		// delete the create authentication step - ignore on failure
		_ = as.steps.Delete(ctx, flow.ID)
		return err
	}
	return nil
}

func (as *Service) assertEmailedCode(
	ctx context.Context,
	assertion authn.Step,
) error {
	// always take the most recent step as the current one - ignore others
	currentStep, err := as.steps.Last(ctx, assertion.IdentityID, assertion.MethodName)
	if err != nil {
		return err
	}
	// check the most recent step has not been already complete
	if currentStep.Complete {
		return merror.Conflict().Describe("emailed code already complete")
	}

	// transform metadata into code metadata structure
	input, err := code.ToMetadata(assertion.RawJSONMetadata)
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

	// try to match codes
	match := stored.Matches(input)
	if !match {
		return merror.Forbidden().From(merror.OriBody).Detail("metadata", merror.DVInvalid)
	}

	// check stored code is not expired
	if time.Now().After(currentStep.CreatedAt.Add(as.codeValidity)) {
		return merror.Forbidden().From(merror.OriBody).Detail("code", merror.DVExpired)
	}

	// complete the authentication step
	return as.steps.CompleteAt(ctx, currentStep.ID, time.Now())
}
