package events

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// DelCounts for couple <identityID, boxID>
func DelCounts(ctx context.Context, redConn *redis.Client, identityID, boxID string) error {
	if _, err := redConn.Del(fmt.Sprintf("%s:%s", identityID, boxID)).Result(); err != nil {
		return err
	}
	return nil
}

// GetCountsForIdentity and return a map with box IDs and their corresponding new events count for the user
func GetCountsForIdentity(ctx context.Context, redConn *redis.Client, identityID string) (map[string]int, error) {
	result := make(map[string]int)
	keys, err := redConn.Keys(fmt.Sprintf("%s:*", identityID)).Result()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return result, nil
	}
	eventsCounts, err := redConn.MGet(keys...).Result()
	if err != nil {
		return nil, err
	}
	for idx, eventCount := range eventsCounts {
		boxID := strings.Split(keys[idx], ":")[1]
		count, err := strconv.Atoi(eventCount.(string))
		if err != nil {
			return nil, merror.Internal().Describef("unexpected value format for %s: %s", keys[idx], err.Error())
		}
		result[boxID] = count
	}
	return result, nil
}

// incrCounts for a given box for all received identityIDs
func incrCounts(ctx context.Context, redConn *redis.Client, identityIDs []string, boxID string) error {
	pipe := redConn.TxPipeline()
	for _, identityID := range identityIDs {
		if _, err := pipe.Incr(fmt.Sprintf("%s:%s", identityID, boxID)).Result(); err != nil {
			return err
		}
	}
	if _, err := pipe.Exec(); err != nil {
		return err
	}
	return nil
}
