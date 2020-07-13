package repositories

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
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
		return nil, merror.NotFound()
	}
	return []byte(value), nil
}
