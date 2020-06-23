package entrypoints

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type IdentityIntraprocessInterface interface {
	Get(ctx context.Context, identityID string) (ret domain.Identity, err error)
	List(ctx context.Context, filters domain.IdentityFilters) ([]*domain.Identity, error)
}

type IdentityIntraprocess struct {
	service identity.IdentityService
}

func NewIdentityIntraprocess(identityService identity.IdentityService) IdentityIntraprocess {
	return IdentityIntraprocess{
		service: identityService,
	}
}

func (intraproc *IdentityIntraprocess) Get(ctx context.Context, identityID string) (ret domain.Identity, err error) {
	return intraproc.service.Get(ctx, identityID)
}

func (intraproc *IdentityIntraprocess) List(ctx context.Context, filters domain.IdentityFilters) ([]*domain.Identity, error) {
	return intraproc.service.List(ctx, filters)
}
