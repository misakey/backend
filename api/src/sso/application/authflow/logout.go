package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// Logout ...
func (afs Service) Logout(ctx context.Context, subject string, token string) error {
	if err := afs.authFlow.DeleteSession(ctx, subject); err != nil {
		return merr.From(err).Desc("delete session")
	}
	if err := afs.authFlow.RevokeToken(ctx, token); err != nil {
		return merr.From(err).Desc("revoke token")
	}
	return nil
}
