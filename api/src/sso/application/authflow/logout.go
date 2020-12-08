package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// Logout ...
func (afs Service) Logout(ctx context.Context, subject string, token string) error {
	if err := afs.authFlow.DeleteSession(ctx, subject); err != nil {
		return merror.Transform(err).Describe("delete session")
	}
	if err := afs.authFlow.RevokeToken(ctx, token); err != nil {
		return merror.Transform(err).Describe("revoke token")
	}
	return nil
}
