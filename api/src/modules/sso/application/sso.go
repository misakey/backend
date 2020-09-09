package application

import (
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/account"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/backuparchive"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/backupkeyshare"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/coupon"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/cryptoaction"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
)

type SSOService struct {
	accountService        account.AccountService
	identityService       identity.IdentityService
	identifierService     identifier.IdentifierService
	authFlowService       authflow.AuthFlowService
	authenticationService authn.Service
	backupKeyShareService backupkeyshare.BackupKeyShareService
	backupArchiveService  backuparchive.BackupArchiveService
	usedCouponService     coupon.UsedCouponService
	cryptoActionService   cryptoaction.CryptoActionService
}

func NewSSOService(
	as account.AccountService, ids identity.IdentityService, idfs identifier.IdentifierService,
	afs authflow.AuthFlowService,
	authns authn.Service,
	bks backupkeyshare.BackupKeyShareService,
	backupArchiveService backuparchive.BackupArchiveService,
	usedCouponService coupon.UsedCouponService,
	cryptoActionService cryptoaction.CryptoActionService,
) SSOService {
	return SSOService{
		accountService: as, identityService: ids, identifierService: idfs,
		authFlowService:       afs,
		authenticationService: authns,
		backupKeyShareService: bks,
		backupArchiveService:  backupArchiveService,
		usedCouponService:     usedCouponService,
		cryptoActionService:   cryptoActionService,
	}
}
