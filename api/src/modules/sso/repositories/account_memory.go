package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type AccountMemory struct {
	byID map[string]domain.Account
}

func NewAccountMemory() *AccountMemory {
	return &AccountMemory{
		byID: make(map[string]domain.Account),
	}
}

func (repo AccountMemory) Create(_ context.Context, account *domain.Account) error {
	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}
	account.ID = id.String()

	fmt.Printf("Insert Account %+v\n", account)
	repo.byID[account.ID] = *account
	return nil
}
