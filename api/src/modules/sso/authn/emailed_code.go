package authn

import (
	"context"
	"fmt"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/authn/code"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

// CreateEmailedCode authentication step
func (as *Service) CreateEmailedCode(ctx context.Context, exec boil.ContextExecutor, identity identity.Identity) error {
	// try to retrieve an existing code for this identity
	existing, err := getLastStep(ctx, exec, identity.ID, oidc.AMREmailedCode)
	if err != nil && !merror.HasCode(err, merror.NotFoundCode) {
		return err
	}
	// if the last authn step is not complete and not expired, we can't create a new one
	if err == nil &&
		!existing.Complete &&
		time.Since(existing.CreatedAt) < as.codeValidity {
		return merror.Conflict().
			Describe("a code has already been generated and not used").
			Detail("identity_id", merror.DVConflict).
			Detail("method_name", merror.DVConflict)
	}

	codeRawJSON, err := code.GenerateAsRawJSON()
	if err != nil {
		return err
	}

	flow := Step{
		IdentityID:      identity.ID,
		MethodName:      oidc.AMREmailedCode,
		RawJSONMetadata: codeRawJSON,

		CreatedAt: time.Now(),

		Complete:   false,
		CompleteAt: null.Time{},
	}
	if err := createStep(ctx, exec, &flow); err != nil {
		return err
	}

	decodedCode, err := code.ToMetadata(codeRawJSON)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"to":   identity.Identifier.Value,
		"code": decodedCode.Code,
	}
	subject := fmt.Sprintf("Votre code de confirmation est %s", decodedCode.Code)
	content, err := as.templates.NewEmail(ctx, identity.Identifier.Value, subject, "code", data)
	if err != nil {
		return err
	}

	if err := as.emails.Send(ctx, content); err != nil {
		return err
	}
	return nil
}

func (as *Service) prepareEmailedCode(
	ctx context.Context, exec boil.ContextExecutor,
	identity identity.Identity, step *Step,
) error {
	step.MethodName = oidc.AMREmailedCode
	// we ignore the conflict error code - if a code already exist, we still want to return authable identity information
	err := as.CreateEmailedCode(ctx, exec, identity)
	// set the error to nil on conflict because we want to fail silently
	// if an emailed code was already generated
	if err != nil && merror.HasCode(err, merror.ConflictCode) {
		return nil
	}
	return err
}

func (as *Service) assertEmailedCode(
	ctx context.Context, exec boil.ContextExecutor,
	assertion Step,
) error {
	// always take the most recent step as the current one - ignore others
	currentStep, err := getLastStep(ctx, exec, assertion.IdentityID, assertion.MethodName)
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
		return merror.Forbidden().From(merror.OriBody).Detail("metadata", merror.DVExpired)
	}

	// complete the authentication step
	return completeAtStep(ctx, exec, currentStep.ID, time.Now())
}
