package events

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func doJoin(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	// check that the current sender is not already a box member
	isMember, err := isMember(ctx, exec, e.BoxID, e.SenderID)
	if err != nil {
		return merror.Transform(err).Describe("checking membership")
	}
	if isMember {
		return merror.Conflict().Describe("already box member")
	}

	// check accesses
	if err := MustHaveAccess(ctx, exec, identities, e.BoxID, e.SenderID); err != nil {
		return merror.Transform(err).Describe("checking accesses")
	}

	return e.persist(ctx, exec)
}

// list active joins for a given box
func listBoxActiveJoinEvents(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]Event, error) {
	// get all the join linked to the box and unreferred
	// refered join event means a leave or kick event has occured. it invalidates them
	activeJoinEvents, err := list(ctx, exec, eventFilters{
		eType:      null.StringFrom(etype.Memberjoin),
		unreferred: true,
		boxID:      null.StringFrom(boxID),
	})
	if err != nil {
		return nil, merror.Transform(err).Describe("listing join events")
	}
	return activeJoinEvents, nil
}

// List box ids joined by an identity ID
func ListMemberBoxIDs(ctx context.Context, exec boil.ContextExecutor, senderID string) ([]string, error) {
	joins, err := list(ctx, exec, eventFilters{
		boxIDOnly:  true,
		eType:      null.StringFrom(etype.Memberjoin),
		unreferred: true,
		senderID:   null.StringFrom(senderID),
		// unkicked:   true,
	})
	if err != nil {
		return nil, err
	}

	joinBoxIDs := make([]string, len(joins))
	for i, e := range joins {
		joinBoxIDs[i] = e.BoxID
	}

	return joinBoxIDs, nil
}
