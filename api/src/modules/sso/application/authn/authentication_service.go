package authn

import (
	"context"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/adaptor/email"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/account"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

type QuotumInterface interface {
	CreateBase(ctx context.Context, identityID string) (interface{}, error)
}

func (s *Service) SetQuotaService(quota QuotumInterface) {
	s.quotaService = quota
}

type Service struct {
	steps             stepRepo
	sessions          sessionRepo
	processes         processRepo
	identifierService identifier.IdentifierService
	identityService   identity.IdentityService
	accountService    account.AccountService
	quotaService      QuotumInterface
	templates         email.Renderer
	emails            email.Sender
	codeValidity      time.Duration
}

type stepRepo interface {
	Create(ctx context.Context, step *Step) error
	CompleteAt(ctx context.Context, stepID int, completeTime time.Time) error
	Last(ctx context.Context, identityID string, methodName oidc.MethodRef) (Step, error)
	Delete(ctx context.Context, stepID int) error
	DeleteIncomplete(ctx context.Context, identityID string) error
}

type sessionRepo interface {
	Upsert(context.Context, Session, time.Duration) error
	Get(context.Context, string) (Session, error)
}

func NewService(
	steps stepRepo, sessions sessionRepo, processes processRepo,
	identifierService identifier.IdentifierService,
	identityService identity.IdentityService,
	accountService account.AccountService,
	templates email.Renderer, emails email.Sender,
) Service {
	return Service{
		steps:             steps,
		sessions:          sessions,
		processes:         processes,
		identifierService: identifierService,
		identityService:   identityService,
		accountService:    accountService,
		templates:         templates,
		emails:            emails,
		codeValidity:      5 * time.Minute,
	}
}
