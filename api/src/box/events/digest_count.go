package events

import (
	"context"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/cache"
)

// IncrDigestCount for a given box for all received identityIDs
func IncrDigestCount(ctx context.Context, redConn *redis.Client, identityIDs []string, boxID string) error {
	pipe := redConn.TxPipeline()
	for _, identityID := range identityIDs {
		if _, err := pipe.Incr(cache.DigestCountKeyByUserBox(identityID, boxID)).Result(); err != nil {
			return err
		}
	}
	if _, err := pipe.Exec(); err != nil {
		return err
	}
	return nil
}

// GetAllDigestCountKeys ...
func GetAllDigestCountKeys(ctx context.Context, redConn *redis.Client) ([]string, error) {
	keys, err := redConn.Keys(cache.DigestCountKeyAll()).Result()
	if err != nil {
		return []string{}, err
	}
	return keys, err
}

// DelDigestCount for couple <identityID, boxID>
func DelDigestCount(ctx context.Context, redConn *redis.Client, identityID, boxID string) error {
	if _, err := redConn.Del(cache.DigestCountKeyByUserBox(identityID, boxID)).Result(); err != nil {
		return err
	}
	return nil
}

// DelAllDigestCountForIdentity identityID
func DelAllDigestCountForIdentity(ctx context.Context, redConn *redis.Client, identityID string) error {
	keys, err := redConn.Keys(cache.DigestCountKeyByUser(identityID)).Result()
	if err != nil {
		return err
	}

	if _, err := redConn.Del(keys...).Result(); err != nil {
		return err
	}
	return nil
}
