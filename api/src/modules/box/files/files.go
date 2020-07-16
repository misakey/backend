package files

import (
	"context"
	"io"
)

type FileRepo interface {
	Upload(context.Context, string, string, io.Reader) error
	Download(context.Context, string, string) ([]byte, error)
	Delete(context.Context, string, string) error
}

func Delete(ctx context.Context, repo FileRepo, boxID, fileID string) error {
	return repo.Delete(ctx, boxID, fileID)
}

func Upload(ctx context.Context, repo FileRepo, boxID, fileID string, encData io.Reader) error {
	return repo.Upload(ctx, boxID, fileID, encData)
}

func Download(ctx context.Context, repo FileRepo, boxID, fileID string) ([]byte, error) {
	return repo.Download(ctx, boxID, fileID)
}
