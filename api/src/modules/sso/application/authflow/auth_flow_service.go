package authflow

import (
	"context"
	"log"
	"net/url"
	"strings"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow/login"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain/consent"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"
)

type AuthFlowService struct {
	identityService identity.IdentityService

	authFlow authFlowRepo

	homePageURL    *url.URL
	loginPageURL   *url.URL
	consentPageURL *url.URL

	selfCliID string
}

func NewAuthFlowService(
	identityService identity.IdentityService,
	authFlow authFlowRepo,
	homePageURI, loginPageURI, consentPageURI, selfCliID string,
) AuthFlowService {
	homePageURL, err := url.ParseRequestURI(homePageURI)
	if err != nil {
		log.Fatalf("invalid home page url: %v", err)
	}
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

		homePageURL:    homePageURL,
		loginPageURL:   loginPageURL,
		consentPageURL: consentPageURL,
		selfCliID:      selfCliID,
	}
}

type authFlowRepo interface {
	GetLoginContext(context.Context, string) (login.Context, error)
	Login(context.Context, string, login.Acceptance) (string, error)

	GetConsentContext(context.Context, string) (consent.Context, error)
	Consent(context.Context, string, consent.Acceptance) (consent.Redirect, error)
	GetConsentSessions(context.Context, string) ([]consent.Session, error)

	DeleteSession(ctx context.Context, subject string) error
	RevokeToken(ctx context.Context, token string) error
}

// returns true if the received string contains `promt=none` string
func HasNonePrompt(authURL string) bool {
	return strings.Contains(authURL, "prompt=none")
}

// returns true if the received string contains no `login_hint=` string
func HasNoLoginHint(authURL string) bool {
	return !strings.Contains(authURL, "login_hint=")
}
