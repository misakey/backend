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

func defaultHandler(ctx context.Context, e *Event, exec boil.ContextExecutor, _ *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	// check that the current sender has access to the box
	if err := MustMemberHaveAccess(ctx, exec, identities, e.BoxID, e.SenderID); err != nil {
		return err
	}

	if e.ReferrerID.Valid {
		return merror.BadRequest().Describe("referrer id cannot be set").Detail("referrer_id", merror.DVForbidden)
	}
	return nil
}

var handlers = map[string]handler{
	"state.lifecycle": lifecycleHandler,

	"msg.text":   defaultHandler,
	"msg.file":   defaultHandler,
	"msg.edit":   defaultHandler,
	"msg.delete": defaultHandler,

	"access.add": addAccessHandler,
	"access.rm":  rmAccessHandler,

	"member.leave": leaveHandler,
	"member.join":  joinHandler,
}

func Handler(eType string) handler {
	return handlers[eType]
}
