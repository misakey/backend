package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v7"
)

// CleanBoxCache ...
func CleanBoxCache(ctx context.Context, redConn *redis.Client, boxID string) error {
	keys, err := redConn.Keys(fmt.Sprintf("*box_%s", boxID)).Result()
	if err != nil {
		return err
	}

	if _, err := redConn.Del(keys...).Result(); err != nil {
		return err
	}

	return nil
}

// CleanBoxesListCache ...
func CleanBoxesListCache(ctx context.Context, redConn *redis.Client, senderID string) error {
	_, err := redConn.Del(GetSenderBoxesKey(senderID)).Result()

	return err
}

// GetSenderBoxesKey ...
func GetSenderBoxesKey(senderID string) string {
	return fmt.Sprintf("cache:user_%s:boxIDs", senderID)
}

// GetBoxMembersKey ...
func GetBoxMembersKey(boxID string) string {
	return fmt.Sprintf("cache:box_%s:membersIDs", boxID)
}

// GetEventCountKey ...
func GetEventCountKey(identityID, boxID string) string {
	return fmt.Sprintf("eventCounts:user_%s:box_%s", identityID, boxID)
}

// GetDigestCountKey ...
func GetDigestCountKey(identityID, boxID string) string {
	return fmt.Sprintf("digestCount:user_%s:box_%s", identityID, boxID)
}

// GetAllDigestCountKeysForIdentity ...
func GetAllDigestCountKeysForIdentity(identityID string) string {
	return fmt.Sprintf("digestCount:user_%s:*", identityID)
}

// GetAllDigestCountKeys ...
func GetAllDigestCountKeys() string {
	return "digestCount:*"
}

// GetEventCountKeys ...
func GetEventCountKeys(identityID string) string {
	return fmt.Sprintf("eventCounts:user_%s:*", identityID)
}
