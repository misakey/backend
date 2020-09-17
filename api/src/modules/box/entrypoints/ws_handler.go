package entrypoints

import (
	"database/sql"

	"github.com/go-redis/redis/v7"
	sentrypoints "gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

type WebsocketHandler struct {
	redConn        *redis.Client
	db             *sql.DB
	identities     sentrypoints.IdentityIntraprocessInterface
	allowedOrigins []string
}

func NewWebsocketHandler(allowedOrigins []string, redConn *redis.Client, db *sql.DB, identities sentrypoints.IdentityIntraprocessInterface) WebsocketHandler {
	return WebsocketHandler{
		allowedOrigins: allowedOrigins,
		redConn:        redConn,
		db:             db,
		identities:     identities,
	}
}
