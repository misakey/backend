package authn

import (
	"context"
	"time"

	"github.com/duo-labs/webauthn/webauthn"

	"gitlab.misakey.dev/misakey/backend/api/src/notifications/email"
)

// QuotumInterface ...
type QuotumInterface interface {
	CreateBase(ctx context.Context, identityID string) (interface{}, error)
}

// Service...
type Service struct {
	sessions  sessionRepo
	processes processRepo

	templates email.Renderer
	emails    email.Sender

	codeValidity time.Duration

	WebauthnHandler *webauthn.WebAuthn
	AppName         string
}

type sessionRepo interface {
	Upsert(context.Context, Session, time.Duration) error
	Get(context.Context, string) (Session, error)
}

// NewService ...
func NewService(
	sessions sessionRepo, processes processRepo,
	templates email.Renderer, emails email.Sender,
	webauthnHandler *webauthn.WebAuthn, appName string,
) Service {
	return Service{
		sessions:        sessions,
		processes:       processes,
		templates:       templates,
		emails:          emails,
		codeValidity:    5 * time.Minute,
		WebauthnHandler: webauthnHandler,
		AppName:         appName,
	}
}
