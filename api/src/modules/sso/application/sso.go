package application

import (
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/account"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authentication"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
)

type SSOService struct {
	accountService        account.AccountService
	identityService       identity.IdentityService
	identifierService     identifier.IdentifierService
	authFlowService       authflow.AuthFlowService
	authenticationService authentication.Service
}

func NewSSOService(
	as account.AccountService, ids identity.IdentityService, idfs identifier.IdentifierService,
	afs authflow.AuthFlowService,
	authns authentication.Service,
) SSOService {
	return SSOService{
		accountService: as, identityService: ids, identifierService: idfs,
		authFlowService:       afs,
		authenticationService: authns,
	}
}
