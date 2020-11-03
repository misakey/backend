package application

import (
	"database/sql"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/account"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/backuparchive"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/backupkeyshare"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/coupon"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/cryptoaction"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/identity"
)

type SSOService struct {
	accountService        account.AccountService
	identityService       identity.IdentityService
	identifierService     identifier.IdentifierService
	authFlowService       authflow.AuthFlowService
	AuthenticationService authn.Service
	backupKeyShareService backupkeyshare.BackupKeyShareService
	backupArchiveService  backuparchive.BackupArchiveService
	usedCouponService     coupon.UsedCouponService
	cryptoActionService   cryptoaction.CryptoActionService

	// NOTE: start to remove repositories components by having storer here (cf box modules)
	sqlDB *sql.DB
}

func NewSSOService(
	as account.AccountService, ids identity.IdentityService, idfs identifier.IdentifierService,
	afs authflow.AuthFlowService,
	authns authn.Service,
	bks backupkeyshare.BackupKeyShareService,
	backupArchiveService backuparchive.BackupArchiveService,
	usedCouponService coupon.UsedCouponService,
	cryptoActionService cryptoaction.CryptoActionService,

	ssoDB *sql.DB,
) SSOService {
	return SSOService{
		accountService: as, identityService: ids, identifierService: idfs,
		authFlowService:       afs,
		AuthenticationService: authns,
		backupKeyShareService: bks,
		backupArchiveService:  backupArchiveService,
		usedCouponService:     usedCouponService,
		cryptoActionService:   cryptoActionService,

		sqlDB: ssoDB,
	}
}
