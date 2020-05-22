package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

type authFlowRepo interface {
	GetLoginContext(ctx context.Context, challenge string) (login.Context, error)
	Login(ctx context.Context, challenge string, acceptance login.Acceptance) (login.Redirect, error)
	// Logout(ctx context.Context, id string) error
	// RevokeToken(ctx context.Context, revocation TokenRevocation) error
}

type Handler struct {
	authFlow authFlowRepo

	loginPageURL string
}

func NewHandler(authFlow authFlowRepo, loginPageURL string) Handler {
	return Handler{
		authFlow:     authFlow,
		loginPageURL: loginPageURL,
	}
}
