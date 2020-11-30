package application

import (
	"database/sql"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/crypto"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/identity"
)

type SSOService struct {
	identityService       identity.IdentityService
	authFlowService       authflow.AuthFlowService
	AuthenticationService authn.Service
	backupKeyShareService crypto.BackupKeyShareService

	// NOTE: start to remove repositories components by having storer here (cf box modules)
	sqlDB   *sql.DB
	redConn *redis.Client
}

func NewSSOService(
	ids identity.IdentityService,
	afs authflow.AuthFlowService,
	authns authn.Service,
	bks crypto.BackupKeyShareService,

	ssoDB *sql.DB,
	redConn *redis.Client,
) SSOService {
	return SSOService{
		identityService:       ids,
		authFlowService:       afs,
		AuthenticationService: authns,
		backupKeyShareService: bks,

		sqlDB:   ssoDB,
		redConn: redConn,
	}
}
