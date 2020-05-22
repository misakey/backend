package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type IdentifierMemory struct {
	byValueKind map[string]domain.Identifier
}

func NewIdentifierMemory() *IdentifierMemory {
	return &IdentifierMemory{
		byValueKind: make(map[string]domain.Identifier),
	}
}

func (repo *IdentifierMemory) Create(_ context.Context, identifier *domain.Identifier) error {
	// generate new UUID for new record
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("could not generate uuid v4")
	}
	identifier.ID = id.String()

	key := identifier.Kind + identifier.Value
	fmt.Printf("Insert Identifier %+v\n", identifier)
	repo.byValueKind[key] = *identifier
	return nil
}

func (repo *IdentifierMemory) GetByKindValue(_ context.Context, kind string, value string) (domain.Identifier, error) {
	key := kind + value
	fmt.Printf("GetByKindValue Identifier %s\n", repo.byValueKind)
	existing, ok := repo.byValueKind[key]
	if !ok {
		return domain.Identifier{}, merror.NotFound().
			Detail("kind", merror.DVNotFound).Detail("value", merror.DVNotFound)
	}
	return existing, nil
}
