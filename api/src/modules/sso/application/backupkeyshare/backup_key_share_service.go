package backupkeyshare

import (
	"context"
	"encoding/json"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
)

type backupKeyShareRepo interface {
	Set(context.Context, string, []byte) error
	Get(context.Context, string) ([]byte, error)
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
	key := backupKeyShare.OtherShareHash
	value, err := json.Marshal(backupKeyShare)
	if err != nil {
		return merror.Internal().Describe("encoding backup key share").Describe(err.Error())
	}
	return bkr.backupKeyShares.Set(ctx, key, value)
}

func (bkr BackupKeyShareService) GetBackupKeyShare(ctx context.Context, otherShareHash string) (*domain.BackupKeyShare, error) {
	value, err := bkr.backupKeyShares.Get(ctx, otherShareHash)
	if err != nil {
		return nil, err
	}
	backupKeyShare := domain.BackupKeyShare{}
	if err := json.Unmarshal(value, &backupKeyShare); err != nil {
		return nil, merror.Internal().Describe("encoding backup key share").Describe(err.Error())
	}
	return &backupKeyShare, nil
}
