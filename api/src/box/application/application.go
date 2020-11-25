package application

import (
	"database/sql"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

type BoxApplication struct {
	DB                *sql.DB
	RedConn           *redis.Client
	identityQuerier   external.IdentityRepo
	cryptoActionsRepo external.CryptoActionRepo
	filesRepo         files.FileStorageRepo
}

func (app *BoxApplication) SetCryptoActionsRepo(repo external.CryptoActionRepo) {
	app.cryptoActionsRepo = repo
}

func NewBoxApplication(db *sql.DB, redConn *redis.Client, filesRepo files.FileStorageRepo) BoxApplication {
	return BoxApplication{
		DB:        db,
		RedConn:   redConn,
		filesRepo: filesRepo,
	}
}

func (app *BoxApplication) SetIdentityRepo(querier external.IdentityRepo) {
	app.identityQuerier = querier
}

func (ba BoxApplication) NewIM() *events.IdentityMapper {
	return events.NewIdentityMapper(ba.identityQuerier)
}
