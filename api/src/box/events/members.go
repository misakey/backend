package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/slice"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
)

// List identities ID that are members of the given box
func ListBoxMemberIDs(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, boxID string) ([]string, error) {
	// 1. try to retrieve cache if it exists
	members, err := redConn.SMembers(cache.GetBoxMembersKey(boxID)).Result()
	if err == nil && len(members) != 0 {
		return members, nil
	}

	// 2. if cache couldn’t be retrieved
	// get the creator id which is a member
	logger.FromCtx(ctx).Debug().Msgf("regenerating members cache for %s", boxID)
	createEvent, err := get(ctx, exec, eventFilters{
		eType: null.StringFrom(etype.Create),
		boxID: null.StringFrom(boxID),
	})
	if err != nil {
		return nil, merror.Transform(err).Describe("getting create event")
	}
	// start the member IDs list with it
	uniqueMemberIDs := make(map[string]bool)
	memberIDs := []string{createEvent.SenderID}
	uniqueMemberIDs[createEvent.SenderID] = true

	// 3. compute people that has joined the box
	activeJoins, err := listBoxActiveJoinEvents(ctx, exec, boxID)
	if err != nil {
		return nil, err
	}

	// build the list and ensure unicity with a map
	for _, e := range activeJoins {
		_, ok := uniqueMemberIDs[e.SenderID]
		if !ok {
			uniqueMemberIDs[e.SenderID] = true
			memberIDs = append(memberIDs, e.SenderID)
		}
	}

	// 4. update the cache
	if _, err := redConn.SAdd(cache.GetBoxMembersKey(boxID), slice.StringSliceToInterfaceSlice(memberIDs)...).Result(); err != nil {
		logger.FromCtx(ctx).Warn().Err(err).Msgf("could not build members cache for %s", boxID)
	}

	return memberIDs, nil
}

func MustBeMember(
	ctx context.Context,
	exec boil.ContextExecutor,
	redConn *redis.Client,
	boxID, senderID string,
) error {

	// check the membership in the cache if it exists
	exists, err := redConn.Exists(cache.GetBoxMembersKey(boxID)).Result()
	if err == nil && exists == 1 {
		// if cache is valid
		senderIsMember, err := redConn.SIsMember(cache.GetBoxMembersKey(boxID), senderID).Result()
		if err != nil {
			if senderIsMember {
				return nil
			}
			return merror.Forbidden().Describe("restricted to member").Detail("sender_id", merror.DVForbidden)
		}
	}

	// if the creator, returns immediatly
	isCreator, err := isCreator(ctx, exec, boxID, senderID)
	if err != nil {
		return err
	}
	if isCreator {
		return nil
	}

	_, err = get(ctx, exec, eventFilters{
		eType:      null.StringFrom("member.join"),
		unreferred: true,
		boxID:      null.StringFrom(boxID),
		// NOTE: today senderID is not used to build unreferred filter since boxID is considered before
		// this is necessary since the sender of member.kick is not the sender of the member.join event.
		senderID: null.StringFrom(senderID),
	})
	// if found, the sender is a member of the box
	if err == nil {
		return nil
	}
	return merror.Forbidden().Describe("restricted to member").Detail("sender_id", merror.DVForbidden)
}

func isMember(
	ctx context.Context,
	exec boil.ContextExecutor,
	redConn *redis.Client,
	boxID,
	senderID string,
) (bool, error) {
	err := MustBeMember(ctx, exec, redConn, boxID, senderID)
	if err != nil && merror.HasCode(err, merror.ForbiddenCode) {
		return false, nil
	}
	// return false admin if an error has occured
	return (err == nil), err
}
