package identity

import (
	"context"
	"database/sql"
)

// AvatarRepo ...
type AvatarRepo interface {
	Upload(context.Context, *AvatarFile) (string, error)
	Delete(context.Context, *AvatarFile) error
}

// Service ...
type Service struct {
	avatars AvatarRepo

	SQLDB *sql.DB
}

// NewService ...
func NewService(
	avatarRepo AvatarRepo,

	ssoDB *sql.DB,
) Service {
	return Service{
		avatars: avatarRepo,

		SQLDB: ssoDB,
	}
}
