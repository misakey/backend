package identity

import (
	"context"
	"io"
)

//
// Models
//

type AvatarFile struct {
	Filename  string
	Extension string

	Data io.Reader
}

//
// Service avatar related methods
//

func (ids IdentityService) UploadAvatar(ctx context.Context, avatar *AvatarFile) (string, error) {
	return ids.avatars.Upload(ctx, avatar)
}

func (ids IdentityService) DeleteAvatar(ctx context.Context, avatar *AvatarFile) error {
	return ids.avatars.Delete(ctx, avatar)
}
