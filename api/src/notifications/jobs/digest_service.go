package jobs

import (
	"database/sql"
	"time"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/notifications/email"
)

// DigestJob contains connectors and configuration for the digest job
type DigestJob struct {
	period    time.Duration
	frequency string
	domain    string
	emails    email.Sender
	templates email.Renderer

	redConn *redis.Client
	boxDB   *sql.DB
	ssoDB   *sql.DB
}

// NewDigestJob constructor
func NewDigestJob(
	frequency, domain string,

	ssoDB *sql.DB,
	boxDB *sql.DB,
	redConn *redis.Client,

	emails email.Sender,
	templates email.Renderer,
) (*DigestJob, error) {
	period, err := GetNotifPeriod(frequency)
	if err != nil {
		return nil, err
	}
	return &DigestJob{
		period:    period,
		frequency: frequency,
		domain:    domain,

		ssoDB:   ssoDB,
		boxDB:   boxDB,
		redConn: redConn,

		emails:    emails,
		templates: templates,
	}, nil
}

// GetNotifPeriod translates a string to a time.Duration
func GetNotifPeriod(frequency string) (time.Duration, error) {
	switch frequency {
	case "minimal":
		return 24 * time.Hour, nil
	case "moderate":
		return 1 * time.Hour, nil
	case "frequent":
		return 5 * time.Minute, nil
	default:
		return 0 * time.Minute, merr.Internal().Descf("wrong frequency value: %s", frequency)
	}
}
