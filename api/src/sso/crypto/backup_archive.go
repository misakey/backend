package crypto

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

// BackupArchive ...
type BackupArchive struct {
	ID          string      `json:"id"`
	AccountID   string      `json:"account_id"`
	Data        null.String `json:"-"`
	CreatedAt   time.Time   `json:"created_at"`
	RecoveredAt null.Time   `json:"recovered_at"`
	DeletedAt   null.Time   `json:"deleted_at"`
}

func newBackupArchive() *BackupArchive { return &BackupArchive{} }

func (b *BackupArchive) fromSQLBoiler(boilModel sqlboiler.BackupArchive) *BackupArchive {
	b.ID = boilModel.ID
	b.AccountID = boilModel.AccountID
	b.Data = boilModel.Data
	b.CreatedAt = boilModel.CreatedAt
	b.RecoveredAt = boilModel.RecoveredAt
	b.DeletedAt = boilModel.DeletedAt
	return b
}

func (b BackupArchive) toSQLBoiler() *sqlboiler.BackupArchive {
	return &sqlboiler.BackupArchive{
		ID:          b.ID,
		AccountID:   b.AccountID,
		Data:        b.Data,
		CreatedAt:   b.CreatedAt,
		RecoveredAt: b.RecoveredAt,
		DeletedAt:   b.DeletedAt,
	}
}

// CreateBackupArchive ...
func CreateBackupArchive(ctx context.Context, exec boil.ContextExecutor, archive BackupArchive) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("generating UUID")
	}
	archive.ID = id.String()
	return archive.toSQLBoiler().Insert(ctx, exec, boil.Infer())
}

// GetBackupArchive ...
func GetBackupArchive(ctx context.Context, exec boil.ContextExecutor, archiveID string) (BackupArchive, error) {
	record, err := sqlboiler.FindBackupArchive(ctx, exec, archiveID)
	if err == sql.ErrNoRows {
		return BackupArchive{}, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	return *newBackupArchive().fromSQLBoiler(*record), err
}

// ListBackupArchives ...
func ListBackupArchives(ctx context.Context, exec boil.ContextExecutor, accountID string) ([]BackupArchive, error) {
	records, err := sqlboiler.BackupArchives(
		// TODO: use blacklisting instead
		// (see https://github.com/volatiletech/sqlboiler/issues/817)
		qm.Select(
			sqlboiler.BackupArchiveColumns.AccountID,
			sqlboiler.BackupArchiveColumns.ID,
			sqlboiler.BackupArchiveColumns.CreatedAt,
			sqlboiler.BackupArchiveColumns.DeletedAt,
			sqlboiler.BackupArchiveColumns.RecoveredAt,
		),
		sqlboiler.BackupArchiveWhere.AccountID.EQ(accountID),
		// most recently created first
		qm.OrderBy(sqlboiler.BackupArchiveColumns.CreatedAt+" desc"),
	).All(ctx, exec)
	domainBackupArchives := make([]BackupArchive, len(records))
	if err == sql.ErrNoRows {
		return domainBackupArchives, nil
	}
	if err != nil {
		return nil, err
	}

	for i, record := range records {
		domainBackupArchives[i] = *newBackupArchive().fromSQLBoiler(*record)
	}
	return domainBackupArchives, nil
}

// GetBackupArchiveMetadata ...
func GetBackupArchiveMetadata(ctx context.Context, exec boil.ContextExecutor, archiveID string) (BackupArchive, error) {
	record, err := sqlboiler.FindBackupArchive(ctx, exec, archiveID,
		sqlboiler.BackupArchiveColumns.AccountID,
		sqlboiler.BackupArchiveColumns.ID,
		sqlboiler.BackupArchiveColumns.CreatedAt,
		sqlboiler.BackupArchiveColumns.DeletedAt,
		sqlboiler.BackupArchiveColumns.RecoveredAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return BackupArchive{}, merror.NotFound().Detail("id", merror.DVNotFound)
		}
		return BackupArchive{}, err
	}
	return *newBackupArchive().fromSQLBoiler(*record), nil
}

// DeleteBackupArchive ...
func DeleteBackupArchive(
	ctx context.Context, exec boil.ContextExecutor,
	archiveID, reason string,
) error {
	deletedAt := null.Time{}
	recoveredAt := null.Time{}

	now := null.TimeFrom(time.Now())
	// TODO: see about using an enum or something like this?
	if reason == "deletion" {
		deletedAt = now
	} else if reason == "recovery" {
		recoveredAt = now
	} else {
		return merror.Internal().Describe("service using deletion reason")
	}

	archive, err := sqlboiler.FindBackupArchive(ctx, exec, archiveID)
	if err != nil {
		if err == sql.ErrNoRows {
			return merror.NotFound().Detail("id", merror.DVNotFound)
		}
		return err
	}

	// it seems that this is how you create null values with "volatiletech/null/v8"
	// (see https://github.com/volatiletech/null/blob/fed49d7/string_test.go#L134)
	archive.Data = null.StringFromPtr(nil)
	archive.DeletedAt = deletedAt
	archive.RecoveredAt = recoveredAt

	rowsAff, err := archive.Update(ctx, exec, boil.Infer())
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Detail("id", merror.DVNotFound).
			Describe("no account rows affected on udpate")
	}
	return nil
}
