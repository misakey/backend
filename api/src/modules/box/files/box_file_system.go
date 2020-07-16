package files

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// BoxFileSystem
type BoxFileSystem struct {
	location string
}

// NewBoxFileSystem's constructor - /!\ not safe to use in production
func NewBoxFileSystem(location string) *BoxFileSystem {
	// create box files directory
	if _, err := os.Stat(location); os.IsNotExist(err) {
		_ = os.Mkdir(location, os.ModePerm)
	}
	return &BoxFileSystem{
		location: location,
	}
}

// getKey by concatenating some info
func (fs *BoxFileSystem) getKey(boxID, fileID string) string {
	return boxID + "_" + fileID
}

// Upload an boxFile in file system directory and return its path
func (fs *BoxFileSystem) Upload(ctx context.Context, boxID, fileID string, data io.Reader,
) error {
	body, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}

	key := fs.getKey(boxID, fileID)
	filePath := path.Join(fs.location, key)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	if _, err := f.Write(body); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Download a box file from local storage and return its raw data
func (fs *BoxFileSystem) Download(ctx context.Context, boxID, fileID string) ([]byte, error) {
	// read the file
	filePath := path.Join(fs.location, fs.getKey(boxID, fileID))
	return ioutil.ReadFile(filePath)
}

// Delete a box file from the file system
func (fs *BoxFileSystem) Delete(ctx context.Context, boxID, fileID string) error {
	path := path.Join(fs.location, fs.getKey(boxID, fileID))
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return merror.NotFound().Describe(err.Error())
		}
		return err
	}
	return os.Remove(path)
}
