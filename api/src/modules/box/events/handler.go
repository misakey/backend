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

func leaveHandler(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client) error {

	// check that the current sender is a box member
	if err := MustBeMember(ctx, exec, e.BoxID, e.SenderID); err != nil {
		// user is a not a box member
		// so we just return
		return nil
	}

	// check that the current sender is not the admin
	// admin can’t leave their own box
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err == nil {
		return merror.Forbidden().Describe("admin can’t leave their own box")
	}

	// get the last join event
	joinEvent, err := GetLastJoin(ctx, exec, e.BoxID, e.SenderID)
	if err != nil {
		return merror.Transform(err).Describe("getting last join event")
	}
	e.RefererID = &joinEvent.ID

	return nil
}

var handlers = map[string]handler{
	"state.lifecycle": adminHandler,
	"member.leave":    leaveHandler,
	"msg.text":        emptyHandler,
	"msg.file":        emptyHandler,
	"msg.edit":        emptyHandler,
	"msg.delete":      emptyHandler,
}

func Handler(eType string) handler {
	return handlers[eType]
}
