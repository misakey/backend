package application

import (
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/account"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authflow"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/authn"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
)

type SSOService struct {
	accountService        account.AccountService
	identityService       identity.IdentityService
	identifierService     identifier.IdentifierService
	authFlowService       authflow.AuthFlowService
	authenticationService authn.Service

	selfCliID string
}

func NewSSOService(
	as account.AccountService, ids identity.IdentityService, idfs identifier.IdentifierService,
	afs authflow.AuthFlowService,
	authns authn.Service,
	selfCliID string,
) SSOService {
	return SSOService{
		accountService: as, identityService: ids, identifierService: idfs,
		authFlowService:       afs,
		authenticationService: authns,
		selfCliID:             selfCliID,
	}
}
