package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/box/external"
	"gitlab.misakey.dev/misakey/backend/api/src/box/files"
)

func doJoin(ctx context.Context, e *Event, _ null.JSON, exec boil.ContextExecutor, redConn *redis.Client, identities *IdentityMapper, _ external.CryptoRepo, _ files.FileStorageRepo) (Metadata, error) {
	// check that the current sender is not already a box member
	isMember, err := isMember(ctx, exec, redConn, e.BoxID, e.SenderID)
	if err != nil {
		return nil, merror.Transform(err).Describe("checking membership")
	}
	if isMember {
		return nil, merror.Conflict().Describe("already box member")
	}

	// check the sender can join the box
	if err := MustBeAbleToJoin(ctx, exec, identities, e.BoxID, e.SenderID); err != nil {
		return nil, merror.Transform(err).Describe("checking joinability")
	}

	identity, err := identities.Get(ctx, e.SenderID, true)
	if err != nil {
		return nil, merror.Transform(err).Describe("doing join")
	}

	// if the sender can join, add it to access.add list if not done yet
	_, err = get(ctx, exec, eventFilters{
		eType:           null.StringFrom(etype.Accessadd),
		unreferred:      true,
		boxID:           null.StringFrom(e.BoxID),
		restrictionType: null.StringFrom("identifier"),
		accessValue:     null.StringFrom(identity.Identifier.Value),
	})
	// NOTE: if the access.add event corresponding to the identifier value is not found, create it
	if err != nil && merror.HasCode(err, merror.NotFoundCode) {
		accessEvent, err := newWithAnyContent(
			etype.Accessadd,
			&accessAddContent{RestrictionType: "identifier", Value: identity.Identifier.Value},
			e.BoxID, e.SenderID, nil,
		)
		if err != nil {
			return nil, merror.Transform(err).Describe("newing a join access.add")
		}
		// persist the generated access.add event
		if err := accessEvent.persist(ctx, exec); err != nil {
			return nil, merror.Transform(err).Describe("persisting a join access.add")
		}
	} else if err != nil {
		return nil, merror.Transform(err).Describe("checking access.add existency")
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
		return nil, merror.Transform(err).Describe("listing join events")
	}
	return activeJoinEvents, nil
}

// ListMemberBoxLatestEvents ...
func ListMemberBoxLatestEvents(ctx context.Context, exec boil.ContextExecutor, senderID string) ([]Event, error) {
	joins, err := list(ctx, exec, eventFilters{
		eType:      null.StringFrom(etype.Memberjoin),
		unreferred: true,
		senderID:   null.StringFrom(senderID),
		// unkicked:   true,
	})
	return joins, err
}
