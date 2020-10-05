package events

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

type Metadata interface{}

type doHandler func(
	context.Context,
	*Event,
	boil.ContextExecutor, // transaction
	*redis.Client,
	entrypoints.IdentityIntraprocessInterface,
	files.FileStorageRepo,
) (Metadata, error)

type afterHandler func(
	context.Context,
	*Event,
	boil.ContextExecutor, // db connector
	*redis.Client,
	entrypoints.IdentityIntraprocessInterface,
	files.FileStorageRepo,
	Metadata,
) error

func empty(_ context.Context, _ *Event, _ boil.ContextExecutor, _ *redis.Client, _ entrypoints.IdentityIntraprocessInterface, _ files.FileStorageRepo) (Metadata, error) {
	return nil, nil
}

// checkReferrer id is set
func checkReferrer(ctx context.Context, e Event) error {
	// check that the current sender has access to the box
	if err := v.ValidateStruct(&e,
		v.Field(&e.ReferrerID, v.Required, is.UUIDv4),
	); err != nil {
		return err
	}

	return nil
}

type EventHandler struct {
	Do    doHandler
	After []afterHandler
}

// NOTE:
// Do handler is at least responsible for making the event persistent in storage (some events might do it differently though).
// After handler must perform non-critical actions that might fail without altering the state of the box.
var eventTypeHandlerMapping = map[string]EventHandler{
	"state.lifecycle": {doLifecycle, gh(publish, notify, interrupt)},

	"msg.text":   {doMessage, gh(publish, notify, computeUsedSpace)},
	"msg.file":   {doMessage, gh(publish, notify, computeUsedSpace)},
	"msg.edit":   {doEditMsg, gh(publish, computeUsedSpace)},
	"msg.delete": {doDeleteMsg, gh(publish, computeUsedSpace)},
	"access.add": {doAddAccess, gh()},
	"access.rm":  {doRmAccess, gh()},

	"member.leave": {doLeave, gh(publish, notify, interrupt, invalidateCaches)},
	"member.join":  {doJoin, gh(publish, notify, invalidateCaches)},
	"member.kick":  {empty, gh(publish, notify, interrupt, invalidateCaches)},
}

func gh(handlers ...afterHandler) []afterHandler {
	return handlers
}

func Handler(eType string) EventHandler {
	return eventTypeHandlerMapping[eType]
}
