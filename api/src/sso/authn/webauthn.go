package authn

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

func prepareWebauthn(
	ctx context.Context, as *Service, exec boil.ContextExecutor,
	curIdentity identity.Identity, currentACR oidc.ClassRef, step *Step,
) (*Step, error) {
	step.MethodName = oidc.AMRWebauthn
	return nil, merr.Forbidden().Desc("not implemented")
}
