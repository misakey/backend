package identity

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type identityRepo interface {
	Create(context.Context, *domain.Identity) error
	Confirm(context.Context, string) error
}

type IdentityService struct {
	identities identityRepo
	proofs     identityProofRepo
}

func NewIdentityService(
	identityRepo identityRepo,
	identityProofRepo identityProofRepo,
) IdentityService {
	return IdentityService{
		identities: identityRepo,
		proofs:     identityProofRepo,
	}
}

func (ids IdentityService) Create(ctx context.Context, identity *domain.Identity) error {
	return ids.identities.Create(ctx, identity)
}

func (ids IdentityService) Confirm(ctx context.Context, id string) error {
	return ids.identities.Confirm(ctx, id)
}
