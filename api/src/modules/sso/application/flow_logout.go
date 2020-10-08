package application

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"
)

// Try to logout the user by invalidating the authentication session
func (sso *SSOService) Logout(ctx context.Context, _ request.Request) (interface{}, error) {
	// verify accesses
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Forbidden()
	}
	if err := sso.authFlowService.Logout(ctx, acc.Subject, acc.Token); err != nil {
		return nil, merror.Transform(err).Describe("logging out on auth")
	}

	// expire all current authentication steps for the logged out subject
	return nil, sso.AuthenticationService.ExpireAll(ctx, acc.IdentityID)
}
