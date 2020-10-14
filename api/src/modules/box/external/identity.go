package external

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type IdentityRepo interface {
	Get(ctx context.Context, identityID string) (domain.Identity, error)
	List(ctx context.Context, filters domain.IdentityFilters) ([]*domain.Identity, error)
}
