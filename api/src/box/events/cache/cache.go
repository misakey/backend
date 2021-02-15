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

// BoxIDsKeyByUserOrg ...
func BoxIDsKeyByUserOrg(senderID string, orgID string) string {
	return fmt.Sprintf("cache:user_%s:org_%s:boxIDs", senderID, orgID)
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

// CleanUserBoxByUserOrg removes cache for a given user/org
// if org id is an empty string, it would flush entirely the user box cache
func CleanUserBoxByUserOrg(
	ctx context.Context, redConn *redis.Client,
	senderID string, orgID string,
) error {
	// set default
	if orgID == "" {
		orgID = "*"
	}

	pattern := fmt.Sprintf("cache:user_%s:org_%s:boxIDs", senderID, orgID)
	keys, err := redConn.Keys(pattern).Result()
	if err != nil {
		return err
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
	return fmt.Sprintf("cache:box_%s:membersIDs", boxID)
}

// CleanBoxByID ...
func CleanBoxByID(ctx context.Context, redConn *redis.Client, boxID string) error {
	keys, err := redConn.Keys(fmt.Sprintf("*box_%s", boxID)).Result()
	if err != nil {
		return err
	}

	if _, err := redConn.Del(keys...).Result(); err != nil {
		return err
	}

	return nil
}
