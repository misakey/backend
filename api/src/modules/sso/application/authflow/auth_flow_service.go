package authflow

import (
	"context"
	"log"
	"net/url"
	"strings"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/consent"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/login"
)

type AuthFlowService struct {
	identityService identity.IdentityService

	authFlow authFlowRepo

	loginPageURL   *url.URL
	consentPageURL *url.URL

	selfCliID string
}

func NewAuthFlowService(
	identityService identity.IdentityService,
	authFlow authFlowRepo,
	loginPageURI string,
	consentPageURI string,
	selfCliID string,
) AuthFlowService {
	loginPageURL, err := url.ParseRequestURI(loginPageURI)
	if err != nil {
		log.Fatalf("invalid login url: %v", err)
	}
	consentPageURL, err := url.ParseRequestURI(consentPageURI)
	if err != nil {
		log.Fatalf("invalid consent url: %v", err)
	}

	return AuthFlowService{
		identityService: identityService,
		authFlow:        authFlow,
		loginPageURL:    loginPageURL,
		consentPageURL:  consentPageURL,
		selfCliID:       selfCliID,
	}
}

type authFlowRepo interface {
	GetLoginContext(context.Context, string) (login.Context, error)
	Login(context.Context, string, login.Acceptance) (login.Redirect, error)

	GetConsentContext(context.Context, string) (consent.Context, error)
	Consent(context.Context, string, consent.Acceptance) (consent.Redirect, error)
	GetConsentSessions(context.Context, string) ([]consent.Session, error)

	DeleteSession(ctx context.Context, subject string) error
	RevokeToken(ctx context.Context, token string) error
}

func NonePrompt(requestURL string) bool {
	return strings.Contains(requestURL, "prompt=none")
}
