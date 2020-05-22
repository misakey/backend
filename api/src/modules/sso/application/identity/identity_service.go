package identity

import (
	"context"

	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type identityRepo interface {
	Create(context.Context, *domain.Identity) error
	Get(context.Context, string) (domain.Identity, error)
	Confirm(context.Context, string) error
	List(context.Context, domain.IdentityFilters) ([]*domain.Identity, error)
}

type IdentityService struct {
	identities identityRepo
}

func NewIdentityService(
	identityRepo identityRepo,
) IdentityService {
	return IdentityService{
		identities: identityRepo,
	}
}

func (ids IdentityService) Create(ctx context.Context, identity *domain.Identity) error {
	return ids.identities.Create(ctx, identity)
}

func (ids IdentityService) Get(ctx context.Context, identityID string) (domain.Identity, error) {
	return ids.identities.Get(ctx, identityID)
}

func (ids IdentityService) GetAuthableByIdentifierID(ctx context.Context, identifierID string) (domain.Identity, error) {
	filters := domain.IdentityFilters{
		IdentifierID: null.StringFrom(identifierID),
		IsAuthable:   null.BoolFrom(true),
	}
	identities, err := ids.identities.List(ctx, filters)
	if err != nil {
		return domain.Identity{}, err
	}
	if len(identities) < 1 {
		return domain.Identity{}, merror.NotFound().
			Detail("identifier_id", merror.DVNotFound).
			Detail("is_authable", merror.DVNotFound)
	}
	if len(identities) > 1 {
		return domain.Identity{}, merror.Internal().Describef("more than one authable identity found for %s", identifierID)
	}
	return *identities[0], nil
}

func (ids IdentityService) Confirm(ctx context.Context, id string) error {
	return ids.identities.Confirm(ctx, id)
}
