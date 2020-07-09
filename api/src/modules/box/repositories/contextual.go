package repositories

import (
	"context"
	"database/sql"
	"io"

	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

// Contextual interface is used to pass contextual repositories between actors
// it allows to use repo transactions accross our logic but also
// to mock layer for unit testing purpose
type Contextual interface {
	DB() boil.ContextExecutor
	Identities() entrypoints.IdentityIntraprocessInterface
	Files() fileRepo
	EventsCounts() eventsCountRepo
}

type fileRepo interface {
	Upload(context.Context, string, string, io.Reader) error
	Download(context.Context, string, string) ([]byte, error)
	Delete(context.Context, string, string) error
}

type eventsCountRepo interface {
	Incr(ctx context.Context, identityIDs []string, boxID string) error
	Del(ctx context.Context, identityID, boxID string) error
	GetIdentityEventsCount(ctx context.Context, identityID string) (map[string]int, error)
}

// RealWorld implementation of the Contextual repository using real infrastructure storage
type RealWorld struct {
	db           *sql.DB
	identityRepo entrypoints.IdentityIntraprocessInterface
	files        fileRepo
	eventsCounts eventsCountRepo
}

func NewRealWorld(
	db *sql.DB,
	identityRepo entrypoints.IdentityIntraprocessInterface,
	files fileRepo,
	eventsCounts eventsCountRepo,
) RealWorld {
	return RealWorld{
		db:           db,
		identityRepo: identityRepo,
		files:        files,
		eventsCounts: eventsCounts,
	}
}

func (r RealWorld) DB() boil.ContextExecutor {
	return r.db
}
func (r RealWorld) Identities() entrypoints.IdentityIntraprocessInterface {
	return r.identityRepo
}
func (r RealWorld) Files() fileRepo {
	return r.files
}
func (r RealWorld) EventsCounts() eventsCountRepo {
	return r.eventsCounts
}
