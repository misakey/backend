package identity

import (
	"context"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type identityRepo interface {
	Create(context.Context, *domain.Identity) error
	Get(context.Context, string) (domain.Identity, error)
	Update(context.Context, *domain.Identity) error
	List(context.Context, domain.IdentityFilters) ([]*domain.Identity, error)
}

type AvatarRepo interface {
	Upload(context.Context, *domain.AvatarFile) (string, error)
	Delete(context.Context, *domain.AvatarFile) error
}

type IdentityService struct {
	identities identityRepo
	avatars    AvatarRepo

	identifierService identifier.IdentifierService
}

func NewIdentityService(
	identityRepo identityRepo,
	avatarRepo AvatarRepo,
	identifierService identifier.IdentifierService,
) IdentityService {
	return IdentityService{
		identities: identityRepo,
		avatars:    avatarRepo,

		identifierService: identifierService,
	}
}

func (ids IdentityService) Create(ctx context.Context, identity *domain.Identity) error {
	if err := ids.identities.Create(ctx, identity); err != nil {
		return merror.Transform(err).Describe("creating identity")
	}
	return nil
}

func (ids IdentityService) Get(ctx context.Context, identityID string) (ret domain.Identity, err error) {
	if ret, err = ids.identities.Get(ctx, identityID); err != nil {
		return ret, merror.Transform(err).Describe("getting identity")
	}

	// retrieve the related identifier
	ret.Identifier, err = ids.identifierService.Get(ctx, ret.IdentifierID)
	if err != nil {
		return ret, merror.Transform(err).Describe("getting identifier")
	}
	return ret, nil
}

func (ids IdentityService) List(ctx context.Context, filters domain.IdentityFilters) ([]*domain.Identity, error) {
	return ids.identities.List(ctx, filters)
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
		return merror.Transform(err).Describe("updating identity")
	}
	return nil
}
