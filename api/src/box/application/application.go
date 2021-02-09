package application

import (
	"database/sql"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

// BoxApplication contains connectors and repositories
// to interact with all box module services
type BoxApplication struct {
	DB        *sql.DB
	RedConn   *redis.Client
	filesRepo files.FileStorageRepo
	selfOrgID string

	identityRepo external.IdentityRepo
	cryptoRepo   external.CryptoRepo
}

// NewBoxApplication constructor
func NewBoxApplication(
	db *sql.DB, redConn *redis.Client,
	filesRepo files.FileStorageRepo,
	selfOrgID string,

	identityRepo external.IdentityRepo,
	cryptoRepo external.CryptoRepo,
) BoxApplication {
	return BoxApplication{
		DB:        db,
		RedConn:   redConn,
		filesRepo: filesRepo,
		selfOrgID: selfOrgID,

		identityRepo: identityRepo,
		cryptoRepo:   cryptoRepo,
	}
}

// NewIM constructor
func (app BoxApplication) NewIM() *events.IdentityMapper {
	return events.NewIdentityMapper(app.identityRepo)
}
