package events

import (
	"context"

	"github.com/go-redis/redis/v7"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/cache"
)

// DelToNotify for couple <identityID, boxID>
func DelToNotify(ctx context.Context, redConn *redis.Client, identityID, boxID string) error {
	if _, err := redConn.Del(cache.GetToNotifyKey(identityID, boxID)).Result(); err != nil {
		return err
	}
	return nil
}

// incrToNotify for a given box for all received identityIDs
func IncrToNotify(ctx context.Context, redConn *redis.Client, identityIDs []string, boxID string) error {
	pipe := redConn.TxPipeline()
	for _, identityID := range identityIDs {
		if _, err := pipe.Incr(cache.GetToNotifyKey(identityID, boxID)).Result(); err != nil {
			return err
		}
	}
	if _, err := pipe.Exec(); err != nil {
		return err
	}
	return nil
}

// GetAllToNotifyKeys
func GetAllToNotifyKeys(ctx context.Context, redConn *redis.Client) ([]string, error) {
	keys, err := redConn.Keys(cache.GetAllToNotifyKeys()).Result()
	if err != nil {
		return []string{}, err
	}
	return keys, err
}

// DelAllToNotifyForIdentity
func DelAllToNotifyForIdentity(ctx context.Context, redConn *redis.Client, identityID string) error {
	keys, err := redConn.Keys(cache.GetAllToNotifyKeysForIdentity(identityID)).Result()
	if err != nil {
		return err
	}

	if _, err := redConn.Del(keys...).Result(); err != nil {
		return err
	}
	return nil
}
