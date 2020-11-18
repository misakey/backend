package identity

import (
	"context"
	"database/sql"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/application/identifier"
)

type AvatarRepo interface {
	Upload(context.Context, *AvatarFile) (string, error)
	Delete(context.Context, *AvatarFile) error
}

type IdentityService struct {
	profileSharingConsents *profileSharingConsentSQLRepo
	avatars                AvatarRepo

	identifierService identifier.IdentifierService

	SqlDB *sql.DB
}

func NewIdentityService(
	profileSharingConsentRepo *profileSharingConsentSQLRepo,
	avatarRepo AvatarRepo,
	identifierService identifier.IdentifierService,

	ssoDB *sql.DB,
) IdentityService {
	return IdentityService{
		profileSharingConsents: profileSharingConsentRepo,
		avatars:                avatarRepo,

		identifierService: identifierService,

		SqlDB: ssoDB,
	}
}
