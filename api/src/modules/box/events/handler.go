package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type handler func(
	context.Context,
	*Event,
	boil.ContextExecutor,
	*redis.Client,
) error

func emptyHandler(_ context.Context, _ *Event, _ boil.ContextExecutor, _ *redis.Client) error {
	return nil
}

func adminHandler(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client) error {
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err != nil {
		return merror.Transform(err).Describe("checking admin")
	}
	return nil
}

var handlers = map[string]handler{
	"state.lifecycle": adminHandler,
	"msg.text":        emptyHandler,
	"msg.file":        emptyHandler,
	"msg.edit":        emptyHandler,
	"msg.delete":      emptyHandler,
}

func Handler(eType string) handler {
	return handlers[eType]
}
