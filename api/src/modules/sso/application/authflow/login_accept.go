package authflow

import (
	"context"

  "gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

// LoginGetContext using a login challenge
func (af *AuthFlowService) LoginGetContext(ctx context.Context, loginChallenge string) (login.Context, error) {
  return af.authFlow.GetLoginContext(ctx, loginChallenge)
}


// Accept the login step
func (af *AuthFlowService) LoginAccept(ctx context.Context, loginChallenge string, acceptance login.Acceptance) (login.Redirect, error) {
  return af.authFlow.Login(ctx, loginChallenge, acceptance)
}
