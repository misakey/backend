package events

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
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

// doReferrer checks that the referrer_id is set
// and **do not** persist event in database
func doReferrer(ctx context.Context, e *Event, exec boil.ContextExecutor, _ *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	// check that the current sender has access to the box
	if err := MustMemberHaveAccess(ctx, exec, identities, e.BoxID, e.SenderID); err != nil {
		return err
	}

	if err := v.ValidateStruct(e,
		v.Field(&e.ReferrerID, v.Required, is.UUIDv4),
	); err != nil {
		return err
	}

	return nil
}

type EventHandler struct {
	Do, After handler
}

// NOTE:
// Do handler is at least responsible for making the event persistent in storage (some events might do it differently though).
// After handler must perform non-critical actions that might fail without altering the state of the box.
var eventTypeHandlerMapping = map[string]EventHandler{
	"state.lifecycle": {doLifecycle, publish},

	"msg.text":   {doDefault, publishAndNotify},
	"msg.file":   {doDefault, publishAndNotify},
	"msg.edit":   {doReferrer, publish},
	"msg.delete": {doReferrer, publish},
	"access.add": {doAddAccess, empty},
	"access.rm":  {doRmAccess, empty},

	"member.leave": {doLeave, publishAndNotify},
	"member.join":  {doJoin, publishAndNotify},
	"member.kick":  {empty, publishAndNotify},
}

func Handler(eType string) EventHandler {
	return eventTypeHandlerMapping[eType]
}
