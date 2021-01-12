package events

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/cache"
)

// DelCounts for couple <identityID, boxID>
func DelCounts(ctx context.Context, redConn *redis.Client, identityID, boxID string) error {
	if _, err := redConn.Del(cache.GetEventCountKey(identityID, boxID)).Result(); err != nil {
		return err
	}
	return nil
}

// GetCountsForIdentity and return a map with box IDs and their corresponding new events count for the user
func GetCountsForIdentity(ctx context.Context, redConn *redis.Client, identityID string) (map[string]int, error) {
	result := make(map[string]int)
	keys, err := redConn.Keys(cache.GetEventCountKeys(identityID)).Result()
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
		boxID := strings.Trim(strings.Split(keys[idx], ":")[2], "box_")
		count, err := strconv.Atoi(eventCount.(string))
		if err != nil {
			return nil, merr.Internal().Descf("unexpected value format for %s: %s", keys[idx], err.Error())
		}
		result[boxID] = count
	}
	return result, nil
}

// GetCountForIdentity and return an int for the asked box
func GetCountForIdentity(ctx context.Context, redConn *redis.Client, identityID, boxID string) (int, error) {
	eventsCount, err := redConn.Get(cache.GetEventCountKey(identityID, boxID)).Int()
	if err != nil && err == redis.Nil {
		// if no result, then there is no new event
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return eventsCount, nil
}

// IncrBoxCounts for a given box for all received identityIDs
func IncrBoxCounts(ctx context.Context, redConn *redis.Client, identityIDs []string, boxID string) error {
	pipe := redConn.TxPipeline()
	for _, identityID := range identityIDs {
		if _, err := pipe.Incr(cache.GetEventCountKey(identityID, boxID)).Result(); err != nil {
			return err
		}
	}
	if _, err := pipe.Exec(); err != nil {
		return err
	}
	return nil
}

// ComputeCount ...
func ComputeCount(ctx context.Context, redConn *redis.Client, senderID, boxID string) int {
	eventsCount, err := GetCountsForIdentity(ctx, redConn, senderID)
	if err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("could not get events count for %s:%s", senderID, boxID)
		return 0
	}
	// if there is no value for a given box
	// that means no new event since last visit
	count, ok := eventsCount[boxID]
	if !ok {
		return 0
	}
	return count
}
