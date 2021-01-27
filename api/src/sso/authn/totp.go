package authn

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

func prepareTOTP(
	ctx context.Context, as *Service, exec boil.ContextExecutor, _ *redis.Client,
	curIdentity identity.Identity, currentACR oidc.ClassRef, step *Step,
) (*Step, error) {
	step.MethodName = oidc.AMRTOTP
	return nil, merr.Forbidden().Desc("not implemented")
}
