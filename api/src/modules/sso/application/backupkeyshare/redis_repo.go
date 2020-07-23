package backupkeyshare

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/domain"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type RedisRepo struct {
	repositories.SimpleKeyRedis

	keyExpiration time.Duration
}

func NewRedisRepo(redConn *redis.Client, keyExpiration time.Duration) *RedisRepo {
	return &RedisRepo{repositories.NewSimpleKeyRedis(redConn), keyExpiration}
}

func (rr RedisRepo) key(otherShareHash string) string {
	return "backupkeyshare:" + otherShareHash
}

func (rr RedisRepo) Insert(ctx context.Context, backupKeyShare domain.BackupKeyShare) error {
	key := rr.key(backupKeyShare.OtherShareHash)
	value, err := json.Marshal(backupKeyShare)
	if err != nil {
		return merror.Internal().Describe("encoding backup key share").Describe(err.Error())
	}
	return rr.SimpleKeyRedis.Set(ctx, key, value, rr.keyExpiration)
}

func (rr RedisRepo) Get(ctx context.Context, otherShareHash string) (domain.BackupKeyShare, error) {
	keyShare := domain.BackupKeyShare{}
	value, err := rr.SimpleKeyRedis.Get(ctx, rr.key(otherShareHash))
	if err != nil {
		return keyShare, err
	}
	if err := json.Unmarshal(value, &keyShare); err != nil {
		return keyShare, merror.Transform(err).Describe("unmarshaling backup key share")
	}
	return keyShare, nil
}
