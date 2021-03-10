package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v7"
)

///
// User (Identity) Cache
// - store box ids per org ids
// - store events counts per box ids
// - store digest count per box ids

// BoxIDsKeyByUserOrgDatatag ...
func BoxIDsKeyByUserOrgDatatag(senderID, orgID, datatagID string) string {
	return fmt.Sprintf("cache:user_%s:org_%s:datatag_%s:boxIDs", senderID, orgID, datatagID)
}

// BoxIDsKeysByUser ...
func BoxIDsKeysByUser(senderID string) string {
	return fmt.Sprintf("cache:user_%s:org_*:datatag_*:boxIDs", senderID)
}

// EventCountKeyByUserBox ...
func EventCountKeyByUserBox(userID, boxID string) string {
	return fmt.Sprintf("eventCounts:user_%s:box_%s", userID, boxID)
}

// EventCountKeyByUser ...
func EventCountKeyByUser(userID string) string {
	return fmt.Sprintf("eventCounts:user_%s:*", userID)
}

// DigestCountKeyByUserBox ...
func DigestCountKeyByUserBox(userID, boxID string) string {
	return fmt.Sprintf("digestCount:user_%s:box_%s", userID, boxID)
}

// DigestCountKeyByUser ...
func DigestCountKeyByUser(userID string) string {
	return fmt.Sprintf("digestCount:user_%s:*", userID)
}

// DigestCountKeyAll ...
func DigestCountKeyAll() string {
	return "digestCount:*"
}

// CleanUserBoxByUser removes cache for a given user
func CleanUserBoxByIdentity(
	ctx context.Context, redConn *redis.Client,
	senderID string,
) error {
	pattern := fmt.Sprintf("cache:user_%s:org_*:datatag_*:boxIDs", senderID)
	keys, err := redConn.Keys(pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	if _, err := redConn.Del(keys...).Result(); err != nil {
		return err
	}
	return nil
}

//
// Box Cache:
// - store member IDs
//

// MemberIDsKeyByBox ...
func MemberIDsKeyByBox(boxID string) string {
	return fmt.Sprintf("cache:box_%s:memberIDs", boxID)
}

// CleanBoxByID ...
func CleanBoxByID(ctx context.Context, redConn *redis.Client, boxID string) error {
	keys, err := redConn.Keys(fmt.Sprintf("*box_%s:*", boxID)).Result()
	if err != nil {
		return err
	}

	if _, err := redConn.Del(keys...).Result(); err != nil {
		return err
	}

	return nil
}

// CleanBoxMembersByID ...
func CleanBoxMembersByID(ctx context.Context, redConn *redis.Client, boxID string) error {
	if _, err := redConn.Del(fmt.Sprintf("cache:box_%s:memberIDs", boxID)).Result(); err != nil {
		return err
	}

	return nil
}
