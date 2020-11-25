package authn

import (
	"context"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/notifications/email"
)

type QuotumInterface interface {
	CreateBase(ctx context.Context, identityID string) (interface{}, error)
}

func (as *Service) SetQuotaService(quota QuotumInterface) {
	as.quotaService = quota
}

type Service struct {
	sessions  sessionRepo
	processes processRepo

	quotaService QuotumInterface

	templates email.Renderer
	emails    email.Sender

	codeValidity time.Duration
}

type sessionRepo interface {
	Upsert(context.Context, Session, time.Duration) error
	Get(context.Context, string) (Session, error)
}

func NewService(
	sessions sessionRepo, processes processRepo,
	templates email.Renderer, emails email.Sender,
) Service {
	return Service{
		sessions:     sessions,
		processes:    processes,
		templates:    templates,
		emails:       emails,
		codeValidity: 5 * time.Minute,
	}
}
