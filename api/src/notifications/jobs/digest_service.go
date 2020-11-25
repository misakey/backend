package jobs

import (
	"database/sql"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/notifications/email"
)

type DigestJob struct {
	period         time.Duration
	frequency      string
	domain         string
	boxExec        boil.ContextExecutor
	redConn        *redis.Client
	identityMapper *events.IdentityMapper
	emails         email.Sender
	templates      email.Renderer

	sqlDB *sql.DB
}

func NewDigestJob(
	frequency,
	domain string,
	boxExec boil.ContextExecutor,
	redConn *redis.Client,
	identityMapper *events.IdentityMapper,
	emails email.Sender,
	templates email.Renderer,
) (*DigestJob, error) {
	period, err := GetNotifPeriod(frequency)
	if err != nil {
		return nil, err
	}
	return &DigestJob{
		period:         period,
		frequency:      frequency,
		domain:         domain,
		boxExec:        boxExec,
		redConn:        redConn,
		identityMapper: identityMapper,
		emails:         emails,
		templates:      templates,
	}, nil
}

func GetNotifPeriod(frequency string) (time.Duration, error) {
	switch frequency {
	case "minimal":
		return 24 * time.Hour, nil
	case "moderate":
		return 1 * time.Hour, nil
	case "frequent":
		return 5 * time.Minute, nil
	default:
		return 0 * time.Minute, merror.Internal().Describef("wrong frequency value: %s", frequency)
	}
}
