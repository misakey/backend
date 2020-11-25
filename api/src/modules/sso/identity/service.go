package identity

import (
	"context"
	"database/sql"
)

type AvatarRepo interface {
	Upload(context.Context, *AvatarFile) (string, error)
	Delete(context.Context, *AvatarFile) error
}

type IdentityService struct {
	avatars AvatarRepo

	SqlDB *sql.DB
}

func NewIdentityService(
	avatarRepo AvatarRepo,

	ssoDB *sql.DB,
) IdentityService {
	return IdentityService{
		avatars: avatarRepo,

		SqlDB: ssoDB,
	}
}
