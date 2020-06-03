package application

import (
	"context"

	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// Try to logout the user by invalidating the authentication session
func (sso SSOService) Logout(ctx context.Context) error {
	// verify accesses
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}
	return sso.authFlowService.Logout(ctx, acc.Subject, acc.Token)
}
