package repositories

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// SimpleKeyRedis manages operations with simple key/value
// storage in redis
type SimpleKeyRedis struct {
	redConn *redis.Client
}

// NewSimpleKeyRedis handles the creation of a SimpleKeyRedis object
func NewSimpleKeyRedis(redConn *redis.Client) SimpleKeyRedis {
	return SimpleKeyRedis{
		redConn: redConn,
	}
}

func (skr *SimpleKeyRedis) Set(ctx context.Context, key string, value []byte, keyExpiration time.Duration) error {
	if _, err := skr.redConn.Set(key, value, keyExpiration).Result(); err != nil {
		return err
	}
	return nil
}

func (skr *SimpleKeyRedis) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := skr.redConn.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, merror.NotFound()
		}
		return nil, err
	}
	return []byte(value), nil
}

func (skr *SimpleKeyRedis) MustFind(ctx context.Context, matchKey string) ([][]byte, error) {
	keys, err := skr.redConn.Keys(matchKey).Result()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, merror.NotFound()
	}

	results, err := skr.redConn.MGet(keys...).Result()
	if err != nil {
		return nil, err
	}

	ret := make([][]byte, len(results))
	for i, elem := range results {
		ret[i] = []byte(elem.(string))
	}
	return ret, nil
}

// Flush the received key without caring about the key existency
func (skr *SimpleKeyRedis) Flush(ctx context.Context, key string) error {
	// NOTE: ignore number of rows remove
	_, err := skr.redConn.Del(key).Result()
	return err
}
