package external

import (
	"context"

	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// IdentityRepo ...
type IdentityRepo interface {
	Get(ctx context.Context, identityID string) (identity.Identity, error)
	List(ctx context.Context, filters identity.Filters) ([]*identity.Identity, error)
	NotificationBulkCreate(ctx context.Context, identityIDs []string, nType string, details null.JSON) error
}
