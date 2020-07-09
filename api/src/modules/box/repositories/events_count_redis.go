package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type EventsCountRedis struct {
	redConn *redis.Client
}

// NewEventsCountRedis init an S3 session
func NewEventsCountRedis(redConn *redis.Client) *EventsCountRedis {
	return &EventsCountRedis{
		redConn: redConn,
	}
}

// Incr events count for a given box
func (ecr *EventsCountRedis) Incr(ctx context.Context, identityIDs []string, boxID string) error {
	// WARNING: here we use a transaction to increment all the keys
	// but this may cause problem with goroutines (several requests for example)
	// so we may want to go deeper in this in order to avoid future problems
	pipe := ecr.redConn.TxPipeline()
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

// Del events count
func (ecr *EventsCountRedis) Del(ctx context.Context, identityID, boxID string) error {
	// we don’t mind if no row is affected
	// so we ignore the first return argument
	if _, err := ecr.redConn.Del(fmt.Sprintf("%s:%s", identityID, boxID)).Result(); err != nil {
		return err
	}
	return nil
}

// GetIdentityEventsCount and return a map with box IDs and their corresponding new events count for the user
func (ecr *EventsCountRedis) GetIdentityEventsCount(ctx context.Context, identityID string) (map[string]int, error) {
	result := make(map[string]int)
	keys, err := ecr.redConn.Keys(fmt.Sprintf("%s:*", identityID)).Result()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return result, nil
	}
	eventsCounts, err := ecr.redConn.MGet(keys...).Result()
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
