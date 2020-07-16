package identities

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func Get(ctx context.Context, iii entrypoints.IdentityIntraprocessInterface, identityID string) (domain.Identity, error) {
	return iii.Get(ctx, identityID)
}
