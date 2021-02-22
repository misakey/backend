package authflow

import (
	"context"
	"log"
	"net/url"
	"strings"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow/consent"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow/login"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow/userinfo"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// Service ...
type Service struct {
	identityService identity.Service

	authFlow authFlowRepo

	homePageURL    *url.URL
	loginPageURL   *url.URL
	consentPageURL *url.URL

	selfCliID string
}

// NewService ...
func NewService(
	identityService identity.Service,
	authFlow authFlowRepo,
	homePageURI, loginPageURI, consentPageURI, selfCliID string,
) Service {
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

	return Service{
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

	GetUserInfo(ctx context.Context, token string) (*userinfo.UserInfo, error)

	CreateClient(ctx context.Context, cli *Client) error
	GetClient(ctx context.Context, id string) (Client, error)
	UpdateClient(ctx context.Context, cli *Client) error
}

// HasNonePrompt returns true if the received string contains `promt=none` string
func HasNonePrompt(authURL string) bool {
	return strings.Contains(authURL, "prompt=none")
}

// HasNoLoginHint returns true if the received string contains no `login_hint=` string
func HasNoLoginHint(authURL string) bool {
	return !strings.Contains(authURL, "login_hint=")
}
