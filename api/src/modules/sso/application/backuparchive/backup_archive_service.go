package backuparchive

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type backupArchiveRepo interface {
	Create(context.Context, *domain.BackupArchive) error
}

type BackupArchiveService struct {
	backupArchives backupArchiveRepo
}

func NewBackupArchiveService(
	repo backupArchiveRepo,
) BackupArchiveService {
	return BackupArchiveService{
		backupArchives: repo,
	}
}

func (service BackupArchiveService) CreateBackupArchive(ctx context.Context, archive domain.BackupArchive) error {
	return service.backupArchives.Create(ctx, &archive)
}
