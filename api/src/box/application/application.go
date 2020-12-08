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
	DB                *sql.DB
	RedConn           *redis.Client
	identityQuerier   external.IdentityRepo
	cryptoActionsRepo external.CryptoActionRepo
	filesRepo         files.FileStorageRepo
}

// SetCryptoActionsRepo to repo
func (app *BoxApplication) SetCryptoActionsRepo(repo external.CryptoActionRepo) {
	app.cryptoActionsRepo = repo
}

// NewBoxApplication constructor
func NewBoxApplication(db *sql.DB, redConn *redis.Client, filesRepo files.FileStorageRepo) BoxApplication {
	return BoxApplication{
		DB:        db,
		RedConn:   redConn,
		filesRepo: filesRepo,
	}
}

// SetIdentityRepo to querier
func (app *BoxApplication) SetIdentityRepo(querier external.IdentityRepo) {
	app.identityQuerier = querier
}

// NewIM constructor
func (app BoxApplication) NewIM() *events.IdentityMapper {
	return events.NewIdentityMapper(app.identityQuerier)
}
