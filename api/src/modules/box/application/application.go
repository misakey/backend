package application

import (
	"database/sql"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

type BoxApplication struct {
	DB         *sql.DB
	RedConn    *redis.Client
	Identities entrypoints.IdentityIntraprocessInterface
	filesRepo  files.FileStorageRepo
}

func (ba *BoxApplication) SetIdentities(identities entrypoints.IdentityIntraprocessInterface) {
	ba.Identities = identities
}

func NewBoxApplication(db *sql.DB, redConn *redis.Client, filesRepo files.FileStorageRepo) BoxApplication {
	return BoxApplication{
		DB:        db,
		RedConn:   redConn,
		filesRepo: filesRepo,
	}
}
