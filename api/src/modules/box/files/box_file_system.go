package files

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// FileSystem
type FileSystem struct {
	location string
}

// NewFileSystem's constructor - /!\ not safe to use in production
func NewFileSystem(location string) *FileSystem {
	// create files directory
	if _, err := os.Stat(location); os.IsNotExist(err) {
		_ = os.Mkdir(location, os.ModePerm)
	}
	return &FileSystem{
		location: location,
	}
}

// Upload an file in file system directory and return its path
func (fs *FileSystem) Upload(ctx context.Context, fileID string, data io.Reader,
) error {
	body, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}

	filePath := path.Join(fs.location, fileID)
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

// Download a file from local storage and return its raw data
func (fs *FileSystem) Download(ctx context.Context, fileID string) ([]byte, error) {
	// read the file
	filePath := path.Join(fs.location, fileID)
	return ioutil.ReadFile(filePath)
}

// Delete a file from the file system
func (fs *FileSystem) Delete(ctx context.Context, fileID string) error {
	path := path.Join(fs.location, fileID)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return merror.NotFound().Describe(err.Error())
		}
		return err
	}
	return os.Remove(path)
}
