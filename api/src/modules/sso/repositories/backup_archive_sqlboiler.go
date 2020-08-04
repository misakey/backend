package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/volatiletech/sqlboiler/boil"
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

// TODO uncomment when it's not unused any more
// func (repo BackupArchiveSQLBoiler) toDomain(boilModel *sqlboiler.BackupArchive) *domain.BackupArchive {
// 	return &domain.BackupArchive{
// 		ID:          boilModel.ID,
// 		AccountID:   boilModel.AccountID,
// 		Data:        boilModel.Data,
// 		CreatedAt:   boilModel.CreatedAt,
// 		RecoveredAt: boilModel.RecoveredAt,
// 		DeletedAt:   boilModel.DeletedAt,
// 	}
// }

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
	id, err := uuid.NewUUID()
	if err != nil {
		return merror.Transform(err).Describe("generating UUID")
	}

	archive.ID = id.String()

	sqlArchive := repo.toSQLBoiler(archive)

	return sqlArchive.Insert(ctx, repo.db, boil.Infer())
}
