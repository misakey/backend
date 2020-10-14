package application

import (
	"database/sql"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

type BoxApplication struct {
	DB              *sql.DB
	RedConn         *redis.Client
	identityQuerier external.IdentityRepo
	filesRepo       files.FileStorageRepo
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
