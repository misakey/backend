package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

type handler func(
	context.Context,
	*Event,
	boil.ContextExecutor,
	*redis.Client,
	entrypoints.IdentityIntraprocessInterface,
) error

func empty(_ context.Context, _ *Event, _ boil.ContextExecutor, _ *redis.Client, _ entrypoints.IdentityIntraprocessInterface) error {
	return nil
}

func doDefault(ctx context.Context, e *Event, exec boil.ContextExecutor, _ *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	// check that the current sender has access to the box
	if err := MustMemberHaveAccess(ctx, exec, identities, e.BoxID, e.SenderID); err != nil {
		return err
	}

	if e.ReferrerID.Valid {
		return merror.BadRequest().Describe("referrer id cannot be set").Detail("referrer_id", merror.DVForbidden)
	}

	return e.persist(ctx, exec)
}

type eventHandler struct {
	Do, After handler
}

// NOTE:
// Do handler is at least responsible for making the event persistent in storage (some events might do it differently though).
// After handler must perform non-critical actions that might fail without altering the state of the box.
var eventTypeHandlerMapping = map[string]eventHandler{
	"state.lifecycle": {doLifecycle, empty},

	"msg.text":   {doDefault, notify},
	"msg.file":   {doDefault, notify},
	"msg.edit":   {empty, empty},
	"msg.delete": {empty, empty},

	"access.add": {doAddAccess, empty},
	"access.rm":  {doRmAccess, empty},

	"member.leave": {doLeave, notify},
	"member.join":  {doJoin, notify},
	"member.kick":  {doJoin, empty},
}

func Handler(eType string) eventHandler {
	return eventTypeHandlerMapping[eType]
}
