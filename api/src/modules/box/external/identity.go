package external

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"
)

type IdentityRepo interface {
	Get(ctx context.Context, identityID string) (identity.Identity, error)
	List(ctx context.Context, filters identity.IdentityFilters) ([]*identity.Identity, error)
}
