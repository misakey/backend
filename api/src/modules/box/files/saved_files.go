package files

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

type SavedFile struct {
	ID                string    `json:"id"`
	IdentityID        string    `json:"identity_id"`
	EncryptedFileID   string    `json:"encrypted_file_id"`
	EncryptedMetadata string    `json:"encrypted_metadata"`
	KeyFingerprint    string    `json:"key_fingerprint"`
	CreatedAt         time.Time `json:"created_at"`
}

func CreateSavedFile(ctx context.Context, exec boil.ContextExecutor, savedFile *SavedFile) error {
	toStore := sqlboiler.SavedFile{
		ID:                savedFile.ID,
		IdentityID:        savedFile.IdentityID,
		EncryptedFileID:   savedFile.EncryptedFileID,
		EncryptedMetadata: savedFile.EncryptedMetadata,
		KeyFingerprint:    savedFile.KeyFingerprint,
	}
	if err := toStore.Insert(ctx, exec, boil.Infer()); err != nil {
		return err
	}
	savedFile.CreatedAt = toStore.CreatedAt
	return nil
}

func DeleteSavedFile(ctx context.Context, exec boil.ContextExecutor, id string) error {
	rowAff, err := sqlboiler.SavedFiles(sqlboiler.SavedFileWhere.ID.EQ(id)).DeleteAll(ctx, exec)
	if err != nil {
		return err
	}
	if rowAff == 0 {
		return merror.NotFound().Detail("id", merror.DVNotFound)
	}
	return nil
}

func CountSavedFilesByIdentityID(ctx context.Context, exec boil.ContextExecutor, identityID string) (int, error) {
	mods := []qm.QueryMod{
		sqlboiler.SavedFileWhere.IdentityID.EQ(identityID),
	}
	count, err := sqlboiler.SavedFiles(mods...).Count(ctx, exec)
	if err != nil {
		return 0, merror.Transform(err).Describe("count save files db")
	}

	return int(count), nil
}

func ListSavedFilesByIdentityID(ctx context.Context, exec boil.ContextExecutor, identityID string, offset *int, limit *int) ([]SavedFile, error) {
	mods := []qm.QueryMod{
		sqlboiler.SavedFileWhere.IdentityID.EQ(identityID),
		qm.OrderBy(sqlboiler.SavedFileColumns.CreatedAt + " DESC"),
	}
	// add offset for pagination
	if offset != nil {
		mods = append(mods, qm.Offset(*offset))
	}
	// add limit for pagination
	if limit != nil {
		mods = append(mods, qm.Limit(*limit))
	}
	dbSavedFiles, err := sqlboiler.SavedFiles(mods...).All(ctx, exec)
	if err == sql.ErrNoRows {
		return []SavedFile{}, nil
	}
	if err != nil {
		return nil, err
	}

	savedFiles := make([]SavedFile, len(dbSavedFiles))
	for idx, savedFile := range dbSavedFiles {
		savedFiles[idx] = toDomain(savedFile)
	}
	return savedFiles, nil
}

func ListSavedFilesByFileID(ctx context.Context, exec boil.ContextExecutor, fileID string) ([]SavedFile, error) {
	dbSavedFiles, err := sqlboiler.SavedFiles(sqlboiler.SavedFileWhere.EncryptedFileID.EQ(fileID)).All(ctx, exec)
	if err == sql.ErrNoRows {
		return []SavedFile{}, nil
	}
	if err != nil {
		return nil, err
	}

	savedFiles := make([]SavedFile, len(dbSavedFiles))
	for idx, savedFile := range dbSavedFiles {
		savedFiles[idx] = toDomain(savedFile)
	}
	return savedFiles, nil
}

func GetSavedFile(ctx context.Context, exec boil.ContextExecutor, id string) (*SavedFile, error) {
	dbSavedFile, err := sqlboiler.SavedFiles(sqlboiler.SavedFileWhere.ID.EQ(id)).One(ctx, exec)
	if err == sql.ErrNoRows {
		return nil, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	if err != nil {
		return nil, err
	}

	savedFile := toDomain(dbSavedFile)

	return &savedFile, nil
}

func toDomain(dbSavedFile *sqlboiler.SavedFile) SavedFile {
	return SavedFile{
		ID:                dbSavedFile.ID,
		IdentityID:        dbSavedFile.IdentityID,
		EncryptedFileID:   dbSavedFile.EncryptedFileID,
		EncryptedMetadata: dbSavedFile.EncryptedMetadata,
		KeyFingerprint:    dbSavedFile.KeyFingerprint,
		CreatedAt:         dbSavedFile.CreatedAt,
	}
}
