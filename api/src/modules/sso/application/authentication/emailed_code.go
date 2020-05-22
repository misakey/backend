package authentication

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/volatiletech/null"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authentication"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// CreateEmailedCode authentication step
func (as *Service) CreateEmailedCode(identityID string) error {
	// TODO10: limit the creation of emailed code in time (one every 5 minutes)
	codeMetadata, err := generateCodeMetadata()
	if err != nil {
		return err
	}

	flow := authentication.Step{
		IdentityID: identityID,
		MethodName: authentication.EmailedCodeMethod,
		Metadata:   codeMetadata,

		InitiatedAt: time.Now(),

		Complete:   false,
		CompleteAt: null.Time{},
	}
	if err := as.steps.Create(&flow); err != nil {
		return err
	}

	// TODO11: send the email
	fmt.Println("Emailed Code: ", string(codeMetadata))
	return nil
}

func (as *Service) assertEmailedCode(currentStep authentication.Step, inputMetadata json.RawMessage) error {
	// transform metadata into code metadata structure
	input, err := toCodeMetadata(inputMetadata)
	if err != nil {
		return merror.Forbidden().From(merror.OriBody).
			Describe(err.Error()).Detail("metadata", merror.DVMalformed)
	}
	stored, err := toCodeMetadata(currentStep.Metadata)
	if err != nil {
		return merror.Forbidden().
			Describef("could not convert step %s as emailed code: %v", currentStep.ID, err.Error()).
			Detail("stored_code", merror.DVMalformed)
	}

	// compare codes
	if input.Code != stored.Code {
		return merror.Forbidden().From(merror.OriBody).Detail("code", merror.DVInvalid)
	}

	// check stored code is not expired
	if time.Now().After(currentStep.InitiatedAt.Add(5 * time.Minute)) {
		return merror.Forbidden().From(merror.OriBody).Detail("code", merror.DVExpired)
	}

	return nil
}
