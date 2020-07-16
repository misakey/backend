package application

import (
	"database/sql"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

type BoxApplication struct {
	db         *sql.DB
	redConn    *redis.Client
	identities entrypoints.IdentityIntraprocessInterface
	filesRepo  files.FileRepo
}

func NewBoxApplication(db *sql.DB, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface, filesRepo files.FileRepo) BoxApplication {
	return BoxApplication{
		db:         db,
		redConn:    redConn,
		identities: identities,
		filesRepo:  filesRepo,
	}
}
