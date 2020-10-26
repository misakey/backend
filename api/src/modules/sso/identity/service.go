package identity

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
)

type AvatarRepo interface {
	Upload(context.Context, *AvatarFile) (string, error)
	Delete(context.Context, *AvatarFile) error
}

type IdentityService struct {
	identities             *identitySQLRepo
	profileSharingConsents *profileSharingConsentSQLRepo
	avatars                AvatarRepo

	identifierService identifier.IdentifierService
}

func NewIdentityService(
	identityRepo *identitySQLRepo,
	profileSharingConsentRepo *profileSharingConsentSQLRepo,
	avatarRepo AvatarRepo,
	identifierService identifier.IdentifierService,
) IdentityService {
	return IdentityService{
		identities:             identityRepo,
		profileSharingConsents: profileSharingConsentRepo,
		avatars:                avatarRepo,

		identifierService: identifierService,
	}
}
