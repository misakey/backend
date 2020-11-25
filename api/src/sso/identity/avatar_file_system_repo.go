package identity

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// avatarFileSystem
type avatarFileSystem struct {
	location  string
	avatarURL string
}

// NewAvatarFileSystem's constructor - /!\ not safe to use in production
func NewAvatarFileSystem(avatarLocation, avatarURL string) *avatarFileSystem {
	// create avatars directory
	if _, err := os.Stat(avatarLocation); os.IsNotExist(err) {
		_ = os.Mkdir(avatarLocation, os.ModePerm)
	}
	return &avatarFileSystem{
		location:  avatarLocation,
		avatarURL: avatarURL,
	}
}

// Upload an avatar in file system directory and return its path
func (fs *avatarFileSystem) Upload(ctx context.Context, avatar *AvatarFile) (string, error) {
	body, err := ioutil.ReadAll(avatar.Data)
	if err != nil {
		return "", err
	}

	filePath := path.Join(fs.location, avatar.Filename)
	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	if _, err := f.Write(body); err != nil {
		f.Close()
		return "", err
	}

	if err := f.Close(); err != nil {
		return "", err
	}

	return fs.avatarURL + "/" + avatar.Filename, nil
}

// Delete an avatar from the file system
func (fs *avatarFileSystem) Delete(ctx context.Context, avatar *AvatarFile) error {
	path := path.Join(fs.location, avatar.Filename)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return merror.NotFound().Describe(err.Error())
		}
		return err
	}

	return os.Remove(path)
}
