package application

import (
	"regexp"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/account"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identity"
)

type SSOService struct {
	accountService    account.AccountService
	identityService   identity.IdentityService
	identifierService identifier.IdentifierService

	displayNameFormat *regexp.Regexp
}

func NewSSOService(
	accountService account.AccountService,
	identityService identity.IdentityService,
	identifierService identifier.IdentifierService,
) SSOService {
	return SSOService{

		accountService:    accountService,
		identityService:   identityService,
		identifierService: identifierService,

		displayNameFormat: regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{1,19}[a-zA-Z0-9]$`),
	}
}
