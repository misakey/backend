package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/authentication"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type AuthenticationStepMemory struct {
	byID         map[string]authentication.Step
	byIdentityID map[string]authentication.Step
}

func NewAuthenticationStepMemory() *AuthenticationStepMemory {
	return &AuthenticationStepMemory{
		byID:         make(map[string]authentication.Step),
		byIdentityID: make(map[string]authentication.Step),
	}
}

func (repo *AuthenticationStepMemory) Create(authStep *authentication.Step) error {
	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}
	authStep.ID = id.String()

	fmt.Printf("Insert Auth Step %+v\n", authStep)
	repo.byID[authStep.ID] = *authStep
	repo.byIdentityID[authStep.IdentityID] = *authStep
	return nil
}

func (repo *AuthenticationStepMemory) Update(ctx context.Context, authStep *authentication.Step) error {
	_, ok := repo.byID[authStep.ID]
	if !ok {
		return merror.NotFound().Detail("id", merror.DVNotFound)
	}

	fmt.Printf("Update Auth Step %+v\n", authStep)
	repo.byID[authStep.ID] = *authStep
	repo.byID[authStep.IdentityID] = *authStep
	return nil
}

func (repo *AuthenticationStepMemory) Last(ctx context.Context, identityID string, methodName authentication.Method) (authentication.Step, error) {
	last := authentication.Step{}
	last, ok := repo.byIdentityID[identityID]
	if !ok {
		return last, merror.NotFound()
	}
	return last, nil
}
