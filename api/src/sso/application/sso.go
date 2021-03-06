package application

import (
	"database/sql"
	"time"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

// SSOService ...
type SSOService struct {
	identityService            identity.Service
	authFlowService            authflow.Service
	AuthenticationService      authn.Service
	backupKeyShareService      crypto.BackupKeyShareService
	rootKeyShareExpirationTime time.Duration
	selfOrgID                  string

	// storers
	ssoDB   *sql.DB
	boxDB   *sql.DB
	redConn *redis.Client
}

// NewSSOService ...
func NewSSOService(
	ids identity.Service,
	afs authflow.Service,
	authns authn.Service,
	bks crypto.BackupKeyShareService,
	rootKeyShareExpirationTime time.Duration,
	selfOrgID string,

	ssoDB, boxDB *sql.DB,
	redConn *redis.Client,
) SSOService {
	return SSOService{
		identityService:            ids,
		authFlowService:            afs,
		AuthenticationService:      authns,
		backupKeyShareService:      bks,
		rootKeyShareExpirationTime: rootKeyShareExpirationTime,
		selfOrgID:                  selfOrgID,

		ssoDB:   ssoDB,
		boxDB:   boxDB,
		redConn: redConn,
	}
}
