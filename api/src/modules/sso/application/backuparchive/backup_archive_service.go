package backuparchive

import (
	"context"
	"time"

	"github.com/volatiletech/null"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type backupArchiveRepo interface {
	Create(context.Context, *domain.BackupArchive) error
	List(context.Context, string) ([]domain.BackupArchive, error)
	GetArchive(context.Context, string) (domain.BackupArchive, error)
	GetArchiveMetadata(context.Context, string) (domain.BackupArchive, error)
	DeleteArchive(ctx context.Context, ID string, deletedAt null.Time, recoveredAt null.Time) error
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

func (service BackupArchiveService) ListBackupArchives(ctx context.Context, accountID string) ([]domain.BackupArchive, error) {
	return service.backupArchives.List(ctx, accountID)
}

func (service BackupArchiveService) GetBackupArchive(ctx context.Context, archiveID string) (domain.BackupArchive, error) {
	return service.backupArchives.GetArchive(ctx, archiveID)
}

func (service BackupArchiveService) GetBackupArchiveMetadata(ctx context.Context, archiveID string) (domain.BackupArchive, error) {
	return service.backupArchives.GetArchiveMetadata(ctx, archiveID)
}

func (service BackupArchiveService) DeleteBackupArchive(ctx context.Context, archiveID string, reason string) error {
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

	return service.backupArchives.DeleteArchive(ctx, archiveID, deletedAt, recoveredAt)
}
