package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v7"
)

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

func CleanBoxesListCache(ctx context.Context, redConn *redis.Client, senderID string) error {
	_, err := redConn.Del(GetSenderBoxesKey(senderID)).Result()

	return err
}

func GetSenderBoxesKey(senderID string) string {
	return fmt.Sprintf("cache:user_%s:boxIDs", senderID)
}

func GetBoxMembersKey(boxID string) string {
	return fmt.Sprintf("cache:box_%s:membersIDs", boxID)
}

func GetEventCountKey(identityID, boxID string) string {
	return fmt.Sprintf("eventCounts:user_%s:box_%s", identityID, boxID)
}

func GetDigestCountKey(identityID, boxID string) string {
	return fmt.Sprintf("digestCount:user_%s:box_%s", identityID, boxID)
}

func GetAllDigestCountKeysForIdentity(identityID string) string {
	return fmt.Sprintf("digestCount:user_%s:*", identityID)
}

func GetAllDigestCountKeys() string {
	return "digestCount:*"
}

func GetEventCountKeys(identityID string) string {
	return fmt.Sprintf("eventCounts:user_%s:*", identityID)
}
