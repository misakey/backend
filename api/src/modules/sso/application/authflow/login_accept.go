package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

// Accept the login step
func (afs AuthFlowService) LoginAccept(ctx context.Context, loginChallenge string, acceptance login.Acceptance) (login.Redirect, error) {
	return afs.authFlow.Login(ctx, loginChallenge, acceptance)
}
