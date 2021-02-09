package authn

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn/code"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

// createEmailedCode authentication step
func (as *Service) createEmailedCode(ctx context.Context, exec boil.ContextExecutor, identity identity.Identity) error {
	// try to retrieve an existing code for this identity
	existing, err := getLastStep(ctx, exec, identity.ID, oidc.AMREmailedCode)
	if err != nil && !merr.IsANotFound(err) {
		return err
	}
	// if the last authn step is not complete and not expired, we can't create a new one
	if err == nil &&
		!existing.Complete &&
		time.Since(existing.CreatedAt) < as.codeValidity {
		return merr.Conflict().
			Desc("a code has already been generated and not used").
			Add("identity_id", merr.DVConflict).
			Add("method_name", merr.DVConflict)
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
		"to":   identity.IdentifierValue,
		"code": decodedCode.Code,
	}
	subject := fmt.Sprintf("Votre code de confirmation est %s", decodedCode.Code)
	content, err := as.templates.NewEmail(ctx, identity.IdentifierValue, subject, "code", data)
	if err != nil {
		return err
	}

	if err := as.emails.Send(ctx, content); err != nil {
		return err
	}
	return nil
}

func prepareEmailedCode(
	ctx context.Context, as *Service, exec boil.ContextExecutor, _ *redis.Client,
	identity identity.Identity, currentACR oidc.ClassRef, step *Step,
	_ bool,
) (*Step, error) {
	step.MethodName = oidc.AMREmailedCode
	// we ignore the conflict error code - if a code already exist, we still want to return identity information
	err := as.createEmailedCode(ctx, exec, identity)
	// set the error to nil on conflict because we want to fail silently
	// if an emailed code was already generated
	if merr.IsAConflict(err) {
		return step, nil
	}
	return step, err
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
		return merr.Conflict().Desc("emailed code already complete")
	}

	// transform metadata into code metadata structure
	input, err := code.ToMetadata(assertion.RawJSONMetadata)
	if err != nil {
		return merr.Forbidden().Ori(merr.OriBody).
			Desc(err.Error()).Add("metadata", merr.DVMalformed)
	}
	stored, err := code.ToMetadata(currentStep.RawJSONMetadata)
	if err != nil {
		return merr.Forbidden().
			Descf("could not convert step %d as emailed code: %v", currentStep.ID, err.Error()).
			Add("stored_code", merr.DVMalformed)
	}

	// try to match codes
	match := stored.Matches(input)
	if !match {
		return merr.Forbidden().Ori(merr.OriBody).Add("metadata", merr.DVInvalid)
	}

	// check stored code is not expired
	if time.Now().After(currentStep.CreatedAt.Add(as.codeValidity)) {
		return merr.Forbidden().Ori(merr.OriBody).Add("metadata", merr.DVExpired)
	}

	// complete the authentication step
	return completeAtStep(ctx, exec, currentStep.ID, time.Now())
}
