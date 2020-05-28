package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

// Accept the login step
func (af AuthFlowService) LoginAccept(ctx context.Context, loginChallenge string, acceptance login.Acceptance) (login.Redirect, error) {
	return af.authFlow.Login(ctx, loginChallenge, acceptance)
}
