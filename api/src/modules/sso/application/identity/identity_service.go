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
	Update(context.Context, *domain.Identity) error
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
	if err := ids.identities.Create(ctx, identity); err != nil {
		return merror.Transform(err).Describe("create identity")
	}
	return nil
}

func (ids IdentityService) Get(ctx context.Context, identityID string) (ret domain.Identity, err error) {
	if ret, err = ids.identities.Get(ctx, identityID); err != nil {
		return ret, merror.Transform(err).Describe("get identity")
	}
	return ret, nil
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

func (ids IdentityService) Update(ctx context.Context, identity *domain.Identity) error {
	if err := ids.identities.Update(ctx, identity); err != nil {
		return merror.Transform(err).Describe("update identity")
	}
	return nil
}
func (ids IdentityService) Confirm(ctx context.Context, id string) error {
	if err := ids.identities.Confirm(ctx, id); err != nil {
		return merror.Transform(err).Describe("confirm identity")
	}
	return nil
}
