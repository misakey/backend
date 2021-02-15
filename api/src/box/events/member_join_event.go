package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

func doJoin(ctx context.Context, e *Event, _ null.JSON, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, _ external.CryptoRepo, _ files.FileStorageRepo) (Metadata, error) {
	// check that the current sender is not already a box member
	isMember, err := isMember(ctx, exec, redConn, e.BoxID, e.SenderID)
	if err != nil {
		return nil, merr.From(err).Desc("checking membership")
	}
	if isMember {
		return nil, merr.Conflict().Desc("already box member")
	}

	// check the sender can join the box
	if err := MustBeAbleToJoin(ctx, exec, identities, e.BoxID, e.SenderID); err != nil {
		return nil, merr.From(err).Desc("checking joinability")
	}

	identity, err := identities.Get(ctx, e.SenderID, true)
	if err != nil {
		return nil, merr.From(err).Desc("doing join")
	}

	// if the sender joins, they must be in the access.add list as identifier so let's verify it is there
	_, err = get(ctx, exec, eventFilters{
		eType:           null.StringFrom(etype.Accessadd),
		unreferred:      true,
		boxID:           null.StringFrom(e.BoxID),
		restrictionType: null.StringFrom("identifier"),
		accessValue:     null.StringFrom(identity.IdentifierValue),
	})
	// NOTE: if the access.add event corresponding to the identifier value is not found, create it
	if merr.IsANotFound(err) {
		accessEvent, newErr := newWithAnyContent(
			etype.Accessadd,
			&accessAddContent{RestrictionType: "identifier", Value: identity.IdentifierValue},
			e.BoxID, e.SenderID, nil,
		)
		if newErr != nil {
			return nil, merr.From(err).Desc("newing a join access.add")
		}
		// persist the generated access.add event
		if err := accessEvent.persist(ctx, exec); err != nil {
			return nil, merr.From(err).Desc("persisting a join access.add")
		}
	} else if err != nil {
		return nil, merr.From(err).Desc("checking access.add existency")
	}
	return nil, e.persist(ctx, exec)
}

// list active joins for a given box
func listBoxActiveJoinEvents(ctx context.Context, exec boil.ContextExecutor, boxID string) ([]Event, error) {
	// get all the join linked to the box and unreferred
	// referred join event means a leave or kick event has occurred. it invalidates them
	activeJoinEvents, err := list(ctx, exec, eventFilters{
		eType:      null.StringFrom(etype.Memberjoin),
		unreferred: true,
		boxID:      null.StringFrom(boxID),
	})
	if err != nil {
		return nil, merr.From(err).Desc("listing join events")
	}
	return activeJoinEvents, nil
}

// ListMemberBoxLatestEvents ...
func ListMemberBoxLatestEvents(ctx context.Context, exec boil.ContextExecutor, senderID string) ([]Event, error) {
	joins, err := list(ctx, exec, eventFilters{
		eType:      null.StringFrom(etype.Memberjoin),
		unreferred: true,
		senderID:   null.StringFrom(senderID),
	})
	return joins, err
}
