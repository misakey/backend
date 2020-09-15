package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func doLeave(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	// check that the current sender has access to the box
	if err := MustMemberHaveAccess(ctx, exec, identities, e.BoxID, e.SenderID); err != nil {
		// user is a not a box member
		// so we just return
		return err
	}

	// check that the current sender is not the admin
	// admin can’t leave their own box
	if err := MustBeAdmin(ctx, exec, e.BoxID, e.SenderID); err == nil {
		return merror.Forbidden().Describe("admin can’t leave their own box")
	}

	// get the last join event to set the referrer id
	joinEvent, err := get(ctx, exec, eventFilters{
		eType:      null.StringFrom("member.join"),
		unreferred: true,
		senderID:   null.StringFrom(e.SenderID),
		boxID:      null.StringFrom(e.BoxID),
	})
	if err != nil {
		return merror.Transform(err).Describe("getting last join event")
	}
	e.ReferrerID = null.StringFrom(joinEvent.ID)

	return e.persist(ctx, exec)
}
