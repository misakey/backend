package events

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/files"
)

func doJoin(ctx context.Context, e *Event, _ null.JSON, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, _ entrypoints.CryptoActionIntraprocessInterface, _ files.FileStorageRepo) (Metadata, error) {
	// check that the current sender is not already a box member
	isMember, err := isMember(ctx, exec, redConn, e.BoxID, e.SenderID)
	if err != nil {
		return nil, merror.Transform(err).Describe("checking membership")
	}
	if isMember {
		return nil, merror.Conflict().Describe("already box member")
	}

	// check accesses
	if err := MustHaveAccess(ctx, exec, identities, e.BoxID, e.SenderID); err != nil {
		return nil, merror.Transform(err).Describe("checking accesses")
	}

	return nil, e.persist(ctx, exec)
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

func ListMemberBoxLatestEvents(ctx context.Context, exec boil.ContextExecutor, senderID string) ([]Event, error) {
	joins, err := list(ctx, exec, eventFilters{
		eType:      null.StringFrom(etype.Memberjoin),
		unreferred: true,
		senderID:   null.StringFrom(senderID),
		// unkicked:   true,
	})
	return joins, err
}
