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

func SendBoxUpdate(ctx context.Context, redConn *redis.Client, memberID string, update *Update) {

	msg, err := update.ToJSON()
	if err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("building box update")
	}
	logger.FromCtx(ctx).Debug().Msgf("send box update to user_%s:ws", memberID)
	if _, err := redConn.Publish(fmt.Sprintf("user_%s:ws", memberID), msg).Result(); err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("sending box update to user_%s:ws", memberID)
	}
}

func SendInterruption(ctx context.Context, redConn *redis.Client, senderID, boxID string) {
	logger.
		FromCtx(ctx).
		Debug().
		Msgf("sending interruption message to %s:%s", boxID, senderID)
	if _, err := redConn.Publish("interrupt:"+boxID+":"+senderID, []byte("stop")).Result(); err != nil {
		logger.FromCtx(ctx).Error().Err(err).Msgf("interrupting channel interrupt:%s:%s", boxID, senderID)
	}
}
