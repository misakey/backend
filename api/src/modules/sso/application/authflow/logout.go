package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

func (af AuthFlowService) Logout(ctx context.Context, subject string, token string) error {
	if err := af.authFlow.DeleteSession(ctx, subject); err != nil {
		return merror.Transform(err).Describe("delete session")
	}
	if err := af.authFlow.RevokeToken(ctx, token); err != nil {
		return merror.Transform(err).Describe("revoke token")
	}
	return nil
}
