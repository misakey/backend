package authflow

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/consent"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

type AuthFlowService struct {
	authFlow authFlowRepo

	loginPageURL   string
	consentPageURL string
}

func NewAuthFlowService(
	authFlow authFlowRepo,
	loginPageURL string,
	consentPageURL string,
) AuthFlowService {
	return AuthFlowService{
		authFlow:       authFlow,
		loginPageURL:   loginPageURL,
		consentPageURL: consentPageURL,
	}
}

type authFlowRepo interface {
	GetLoginContext(context.Context, string) (login.Context, error)
	Login(context.Context, string, login.Acceptance) (login.Redirect, error)

	GetConsentContext(context.Context, string) (consent.Context, error)
	Consent(context.Context, string, consent.Acceptance) (consent.Redirect, error)

	// Logout(ctx context.Context, id string) error
	// RevokeToken(ctx context.Context, revocation TokenRevocation) error
}
