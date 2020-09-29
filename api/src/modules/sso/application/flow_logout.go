package application

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// Try to logout the user by invalidating the authentication session
func (sso SSOService) Logout(ctx context.Context) error {
	// verify accesses
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}
	if err := sso.authFlowService.Logout(ctx, acc.Subject, acc.Token); err != nil {
		return merror.Transform(err).Describe("logging out on auth")
	}

	// expire all current authentication steps for the logged out subject
	return sso.AuthenticationService.ExpireAll(ctx, acc.IdentityID)
}
