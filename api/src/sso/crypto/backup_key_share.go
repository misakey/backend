package crypto

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories"
)

// BackupKeyShare ...
type BackupKeyShare struct {
	AccountID      string `json:"account_id"`
	SaltBase64     string `json:"salt_base64"`
	Share          string `json:"share"`
	OtherShareHash string `json:"other_share_hash"`
}

// BackupKeyShareService ...
type BackupKeyShareService struct {
	repositories.SimpleKeyRedis

	keyExpiration time.Duration
}

// NewBackupKeyShareService ...
func NewBackupKeyShareService(redConn *redis.Client, keyExpiration time.Duration) BackupKeyShareService {
	return BackupKeyShareService{repositories.NewSimpleKeyRedis(redConn), keyExpiration}
}

func (bkr BackupKeyShareService) storageKey(otherShareHash string) string {
	return "backupkeyshare:" + otherShareHash
}

// CreateBackupKeyShare ...
func (bkr BackupKeyShareService) CreateBackupKeyShare(ctx context.Context, backupKeyShare BackupKeyShare) error {
	key := bkr.storageKey(backupKeyShare.OtherShareHash)
	value, err := json.Marshal(backupKeyShare)
	if err != nil {
		return merr.Internal().Desc("encoding backup key share").Desc(err.Error())
	}
	return bkr.SimpleKeyRedis.Set(ctx, key, value, bkr.keyExpiration)
}

// GetBackupKeyShare ...
func (bkr BackupKeyShareService) GetBackupKeyShare(ctx context.Context, otherShareHash string) (*BackupKeyShare, error) {
	keyShare := BackupKeyShare{}
	value, err := bkr.SimpleKeyRedis.Get(ctx, bkr.storageKey(otherShareHash))
	if err != nil {
		return &keyShare, err
	}
	if err := json.Unmarshal(value, &keyShare); err != nil {
		return &keyShare, merr.From(err).Desc("unmarshaling backup key share")
	}
	return &keyShare, nil
}
