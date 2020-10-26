package entrypoints

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"
)

type IdentityIntraprocess struct {
	service identity.IdentityService
}

func NewIdentityIntraprocess(identityService identity.IdentityService) IdentityIntraprocess {
	return IdentityIntraprocess{
		service: identityService,
	}
}

func (intraproc IdentityIntraprocess) Get(ctx context.Context, identityID string) (ret identity.Identity, err error) {
	return intraproc.service.Get(ctx, identityID)
}

func (intraproc IdentityIntraprocess) List(ctx context.Context, filters identity.IdentityFilters) ([]*identity.Identity, error) {
	return intraproc.service.List(ctx, filters)
}
