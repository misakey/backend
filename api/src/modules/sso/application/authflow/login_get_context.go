package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

// LoginGetContext using a login challenge
func (afs AuthFlowService) LoginGetContext(ctx context.Context, loginChallenge string) (login.Context, error) {
	return afs.authFlow.GetLoginContext(ctx, loginChallenge)
}
