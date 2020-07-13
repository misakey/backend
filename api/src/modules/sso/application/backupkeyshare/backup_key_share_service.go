package backupkeyshare

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type backupKeyShareRepo interface {
	Insert(context.Context, domain.BackupKeyShare) error
	Get(context.Context, string) (domain.BackupKeyShare, error)
}

type BackupKeyShareService struct {
	backupKeyShares backupKeyShareRepo
}

func NewBackupKeyShareService(
	backupKeyShareRepo backupKeyShareRepo,
) BackupKeyShareService {
	return BackupKeyShareService{
		backupKeyShares: backupKeyShareRepo,
	}
}

func (bkr BackupKeyShareService) CreateBackupKeyShare(ctx context.Context, backupKeyShare domain.BackupKeyShare) error {
	return bkr.backupKeyShares.Insert(ctx, backupKeyShare)
}

func (bkr BackupKeyShareService) GetBackupKeyShare(ctx context.Context, otherShareHash string) (*domain.BackupKeyShare, error) {
	backupKeyShare, err := bkr.backupKeyShares.Get(ctx, otherShareHash)
	return &backupKeyShare, err
}
