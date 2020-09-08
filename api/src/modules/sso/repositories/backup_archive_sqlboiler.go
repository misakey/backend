package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type BackupArchiveSQLBoiler struct {
	db *sql.DB
}

func NewBackupArchiveSQLBoiler(db *sql.DB) *BackupArchiveSQLBoiler {
	return &BackupArchiveSQLBoiler{
		db: db,
	}
}

func (repo BackupArchiveSQLBoiler) toDomain(boilModel *sqlboiler.BackupArchive) domain.BackupArchive {
	return domain.BackupArchive{
		ID:          boilModel.ID,
		AccountID:   boilModel.AccountID,
		Data:        boilModel.Data,
		CreatedAt:   boilModel.CreatedAt,
		RecoveredAt: boilModel.RecoveredAt,
		DeletedAt:   boilModel.DeletedAt,
	}
}

func (repo BackupArchiveSQLBoiler) toSQLBoiler(domModel *domain.BackupArchive) *sqlboiler.BackupArchive {
	return &sqlboiler.BackupArchive{
		ID:          domModel.ID,
		AccountID:   domModel.AccountID,
		Data:        domModel.Data,
		CreatedAt:   domModel.CreatedAt,
		RecoveredAt: domModel.RecoveredAt,
		DeletedAt:   domModel.DeletedAt,
	}
}

func (repo BackupArchiveSQLBoiler) Create(ctx context.Context, archive *domain.BackupArchive) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return merror.Transform(err).Describe("generating UUID")
	}

	archive.ID = id.String()

	sqlArchive := repo.toSQLBoiler(archive)

	return sqlArchive.Insert(ctx, repo.db, boil.Infer())
}

func (repo BackupArchiveSQLBoiler) List(ctx context.Context, accountID string) ([]domain.BackupArchive, error) {
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
	).All(ctx, repo.db)
	domainBackupArchives := make([]domain.BackupArchive, len(records))
	if err == sql.ErrNoRows {
		return domainBackupArchives, nil
	}
	if err != nil {
		return nil, err
	}

	for i, record := range records {
		domainBackupArchives[i] = repo.toDomain(record)
	}
	return domainBackupArchives, nil
}

func (repo BackupArchiveSQLBoiler) GetArchive(ctx context.Context, ID string) (domain.BackupArchive, error) {
	archive, err := sqlboiler.FindBackupArchive(ctx, repo.db, ID)
	if err == sql.ErrNoRows {
		return domain.BackupArchive{}, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	return repo.toDomain(archive), err
}

func (repo BackupArchiveSQLBoiler) GetArchiveMetadata(ctx context.Context, ID string) (domain.BackupArchive, error) {
	archive, err := sqlboiler.FindBackupArchive(ctx, repo.db, ID,
		sqlboiler.BackupArchiveColumns.AccountID,
		sqlboiler.BackupArchiveColumns.ID,
		sqlboiler.BackupArchiveColumns.CreatedAt,
		sqlboiler.BackupArchiveColumns.DeletedAt,
		sqlboiler.BackupArchiveColumns.RecoveredAt,
	)
	if err == sql.ErrNoRows {
		return domain.BackupArchive{}, merror.NotFound().Detail("id", merror.DVNotFound)
	}
	return repo.toDomain(archive), err
}

func (repo BackupArchiveSQLBoiler) DeleteArchive(ctx context.Context, ID string, deletedAt null.Time, recoveredAt null.Time) error {
	archive, err := sqlboiler.FindBackupArchive(ctx, repo.db, ID)
	if err == sql.ErrNoRows {
		return merror.NotFound().Detail("id", merror.DVNotFound)
	}

	if err != nil {
		return err
	}

	// it seems that this is how you create null values with "volatiletech/null"
	// (see https://github.com/volatiletech/null/blob/fed49d7/string_test.go#L134)
	archive.Data = null.StringFromPtr(nil)

	archive.DeletedAt = deletedAt
	archive.RecoveredAt = recoveredAt

	rowsAff, err := archive.Update(ctx, repo.db, boil.Infer())
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Detail("id", merror.DVNotFound).
			Describe("no account rows affected on udpate")
	}
	return nil
}
