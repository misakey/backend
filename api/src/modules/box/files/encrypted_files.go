package files

import (
	"context"
	"database/sql"
	"io"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

type EncryptedFile struct {
	ID   string
	Size int64
}

type FileStorageRepo interface {
	Upload(context.Context, string, io.Reader) error
	Download(context.Context, string) (io.Reader, error)
	Delete(context.Context, string) error
}

func Create(ctx context.Context, exec boil.ContextExecutor, encryptedFile EncryptedFile) error {
	toStore := sqlboiler.EncryptedFile{
		ID:   encryptedFile.ID,
		Size: encryptedFile.Size,
	}
	return toStore.Insert(ctx, exec, boil.Infer())
}

func Get(ctx context.Context, exec boil.ContextExecutor, fileID string) (*EncryptedFile, error) {
	dbEncryptedFile, err := sqlboiler.EncryptedFiles(sqlboiler.EncryptedFileWhere.ID.EQ(fileID)).One(ctx, exec)
	if err == sql.ErrNoRows {
		return nil, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return nil, err
	}

	encryptedFile := EncryptedFile{
		ID:   dbEncryptedFile.ID,
		Size: dbEncryptedFile.Size,
	}

	return &encryptedFile, nil
}

func Upload(ctx context.Context, repo FileStorageRepo, fileID string, encData io.Reader) error {
	return repo.Upload(ctx, fileID, encData)
}

func Download(ctx context.Context, repo FileStorageRepo, fileID string) (io.Reader, error) {
	return repo.Download(ctx, fileID)
}

func Delete(ctx context.Context, exec boil.ContextExecutor, repo FileStorageRepo, fileID string) error {
	// delete the stored file
	if err := repo.Delete(ctx, fileID); err != nil {
		return err
	}

	// delete file entity (ignoring the no row affected error)
	if _, err := sqlboiler.EncryptedFiles(sqlboiler.EncryptedFileWhere.ID.EQ(fileID)).DeleteAll(ctx, exec); err != nil {
		return err
	}

	return nil
}
