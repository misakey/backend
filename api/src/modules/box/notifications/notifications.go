package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v7"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
)

type Update struct {
	Type   string      `json:"type"`
	Object interface{} `json:"object"`
}

func (u *Update) ToJSON() ([]byte, error) {
	value, err := json.Marshal(u)
	if err != nil {
		return []byte{}, err
	}
	return value, nil
}

func SendUpdate(ctx context.Context, redConn *redis.Client, memberID string, update *Update) {

	msg, err := update.ToJSON()
	if err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("building update")
	}
	logger.FromCtx(ctx).Debug().Msgf("send update to user_%s:ws", memberID)
	if _, err := redConn.Publish(fmt.Sprintf("user_%s:ws", memberID), msg).Result(); err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("sending update to user_%s:ws", memberID)
	}
}
