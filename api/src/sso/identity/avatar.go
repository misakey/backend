package identity

import (
	"context"
	"io"
)

// AvatarFile ...
type AvatarFile struct {
	Filename  string
	Extension string

	Data io.Reader
}

// UploadAvatar ...
func (ids Service) UploadAvatar(ctx context.Context, avatar *AvatarFile) (string, error) {
	return ids.avatars.Upload(ctx, avatar)
}

// DeleteAvatar ...
func (ids Service) DeleteAvatar(ctx context.Context, avatar *AvatarFile) error {
	return ids.avatars.Delete(ctx, avatar)
}
