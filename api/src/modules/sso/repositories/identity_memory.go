package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type IdentityMemory struct {
	byID        map[string]domain.Identity
	byAccountID map[string]domain.Identity
}

func NewIdentityMemory() *IdentityMemory {
	return &IdentityMemory{
		byID:        make(map[string]domain.Identity),
		byAccountID: make(map[string]domain.Identity),
	}
}

func (repo *IdentityMemory) Create(_ context.Context, identity *domain.Identity) error {
	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}
	identity.ID = id.String()

	fmt.Printf("Insert Identity %+v\n", identity)
	repo.byID[identity.ID] = *identity
	repo.byAccountID[identity.AccountID] = *identity
	return nil
}

func (repo *IdentityMemory) Confirm(_ context.Context, identityID string) error {
	// try to get identity
	identity, ok := repo.byID[identityID]
	if !ok {
		return merror.NotFound().Detail("id", merror.DVNotFound)
	}

	identity.Confirmed = true

	fmt.Printf("Confirm Identity %+v\n", identity)
	repo.byID[identity.ID] = identity
	repo.byAccountID[identity.AccountID] = identity
	return nil
}
