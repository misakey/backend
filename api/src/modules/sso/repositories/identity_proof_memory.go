package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type IdentityProofMemory struct {
	byID         map[string]domain.IdentityProof
	byIdentityID map[string]domain.IdentityProof
}

func NewIdentityProofMemory() *IdentityProofMemory {
	return &IdentityProofMemory{
		byID:         make(map[string]domain.IdentityProof),
		byIdentityID: make(map[string]domain.IdentityProof),
	}
}

func (repo *IdentityProofMemory) Create(identityProof *domain.IdentityProof) error {
	// generate new UUID
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}
	identityProof.ID = id.String()

	fmt.Printf("Insert Identity Proof %+v\n", identityProof)
	repo.byID[identityProof.ID] = *identityProof
	repo.byIdentityID[identityProof.IdentityID] = *identityProof
	return nil
}

func (repo *IdentityProofMemory) Update(ctx context.Context, proof *domain.IdentityProof) error {
	_, ok := repo.byID[proof.ID]
	if !ok {
		return merror.NotFound().Detail("id", merror.DVNotFound)
	}

	fmt.Printf("Update Identity Proof %+v\n", proof)
	repo.byID[proof.ID] = *proof
	repo.byID[proof.IdentityID] = *proof
	return nil
}

func (repo *IdentityProofMemory) List(ctx context.Context, filters domain.IdentityProofFilters) ([]*domain.IdentityProof, error) {
	list := []*domain.IdentityProof{}

	fmt.Printf("not fully implemented - List Identity Proof%+v\n", filters)
	if filters.IdentityID != nil {
		existing, ok := repo.byIdentityID[*filters.IdentityID]
		if !ok {
			return list, nil
		}
		list = append(list, &existing)
	}
	return list, nil
}
