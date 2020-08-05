package files

import (
	"context"
	"io"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type FileRepo interface {
	Upload(context.Context, string, string, io.Reader) error
	Download(context.Context, string, string) ([]byte, error)
	DeleteAll(context.Context, string) error
	Delete(context.Context, string, string) error
}

func Upload(ctx context.Context, repo FileRepo, boxID, fileID string, encData io.Reader) error {
	return repo.Upload(ctx, boxID, fileID, encData)
}

func Download(ctx context.Context, repo FileRepo, boxID, fileID string) ([]byte, error) {
	return repo.Download(ctx, boxID, fileID)
}

func EmptyForBox(ctx context.Context, repo FileRepo, boxID string) error {
	err := repo.DeleteAll(ctx, boxID)
	// ignore not found error since the emptiness is satisfied if no record has been found
	if merror.HasCode(err, merror.NotFoundCode) {
		return nil
	}
	return err
}

func Delete(ctx context.Context, repo FileRepo, boxID, fileID string) error {
	return repo.Delete(ctx, boxID, fileID)
}
