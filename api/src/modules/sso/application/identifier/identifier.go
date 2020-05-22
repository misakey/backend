package identifier

import (
	"context"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type identifierRepo interface {
	Create(context.Context, *domain.Identifier) error
	GetByKindValue(context.Context, string, string) (domain.Identifier, error)
}

type IdentifierService struct {
	identifiers identifierRepo
}

func NewIdentifierService(
	identifierRepo identifierRepo,
) IdentifierService {
	return IdentifierService{
		identifiers: identifierRepo,
	}
}

func (ids IdentifierService) EnsureIdentifierExistence(ctx context.Context, identifier *domain.Identifier) error {
	existing, err := ids.identifiers.GetByKindValue(ctx, identifier.Kind, identifier.Value)
	if err != nil && !merror.HasCode(err, merror.NotFoundCode) {
		return err
	}
	// if the identifier has been found, we bind received pointer and return no error
	if err == nil {
		*identifier = existing
		return nil
	}

	// otherwise at this point we know we got a not found error so we create the identifier
	return ids.identifiers.Create(ctx, identifier)
}
