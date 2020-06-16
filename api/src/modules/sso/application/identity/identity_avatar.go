package identity

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

func (ids IdentityService) UploadAvatar(ctx context.Context, avatar *domain.AvatarFile) (string, error) {
	return ids.avatars.Upload(ctx, avatar)
}

func (ids IdentityService) DeleteAvatar(ctx context.Context, avatar *domain.AvatarFile) error {
	return ids.avatars.Delete(ctx, avatar)
}
