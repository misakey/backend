package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/authz"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/slice"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
)

// ListBoxMemberIDs and return their identities ID
func ListBoxMemberIDs(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client,
	boxID string,
) ([]string, error) {
	// 1. try to retrieve cache if it exists
	cacheKey := cache.MemberIDsKeyByBox(boxID)
	memberIDs, err := redConn.SMembers(cacheKey).Result()
	if err == nil && len(memberIDs) > 0 {
		return memberIDs, nil
	}

	// 2. if cache couldn’t be retrieved
	// get the creator id which is a member
	logger.FromCtx(ctx).Debug().Msgf("regenerating members cache for %s", boxID)
	createEvent, err := get(ctx, exec, eventFilters{
		eType: null.StringFrom(etype.Create),
		boxID: null.StringFrom(boxID),
	})
	if err != nil {
		return nil, merr.From(err).Desc("getting create event")
	}
	// start the member IDs list with it
	uniqueMemberIDs := make(map[string]bool)
	memberIDs = []string{createEvent.SenderID}
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
	if _, err := redConn.SAdd(cacheKey, slice.StringSliceToInterfaceSlice(memberIDs)...).Result(); err != nil {
		logger.FromCtx(ctx).Warn().Err(err).Msgf("could not build members cache for %s", boxID)
	}

	return memberIDs, nil
}

func MustBeMemberOrOrg(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client,
	boxID, identityID string,
) error {
	ok, err := isMember(ctx, exec, redConn, boxID, identityID)
	if err != nil {
		return merr.From(err).Desc("checking membership")
	}
	if ok { // if a member, return
		return nil
	}
	// check machine-org case
	acc := oidc.GetAccesses(ctx)
	if acc == nil || authz.IsNotAMachine(*acc) {
		return merr.Forbidden().Desc("nor a member or org")
	}

	// check the connected machine is the org of the box
	createInfo, err := GetCreateInfo(ctx, exec, boxID)
	if err != nil {
		return merr.From(err).Desc("getting create info")
	}
	if identityID != createInfo.OwnerOrgID {
		return merr.Forbidden().Desc("nor a member or org")
	}
	return nil
}

// MustBeMember ...
func MustBeMember(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client,
	boxID, identityID string,
) error {
	// check the membership in the cache if it exists
	cacheKey := cache.MemberIDsKeyByBox(boxID)
	exists, err := redConn.Exists(cacheKey).Result()
	if err == nil && exists == 1 {
		// if cache is valid
		senderIsMember, err := redConn.SIsMember(cacheKey, identityID).Result()
		if err != nil {
			if senderIsMember {
				return nil
			}
			return merr.Forbidden().Desc("must be a member").Add("reason", "not_member").Add("sender_id", merr.DVForbidden)
		}
	}

	// get member.join events that are not referred (kicks/leaves...)
	_, err = get(ctx, exec, eventFilters{
		eType:    null.StringFrom(etype.Memberjoin),
		boxID:    null.StringFrom(boxID),
		senderID: null.StringFrom(identityID),
		// exclude referred member join events
		excludeOnRef: &referentsFilters{
			eTypes:   []string{etype.Memberleave, etype.Memberkick},
			senderID: null.StringFrom(identityID),
			boxID:    null.StringFrom(boxID),
		},
	})
	// if found, the sender is a member of the box
	if err == nil {
		return nil
	}

	// if the creator, it is a member of the box
	isCreator, err := isCreator(ctx, exec, boxID, identityID)
	if err != nil {
		return err
	}
	if isCreator {
		return nil
	}

	return merr.Forbidden().Desc("must be a member").Add("reason", "not_member").Add("sender_id", merr.DVForbidden)
}

func isMember(
	ctx context.Context,
	exec boil.ContextExecutor, redConn *redis.Client,
	boxID, senderID string,
) (bool, error) {
	err := MustBeMember(ctx, exec, redConn, boxID, senderID)
	if merr.IsAForbidden(err) {
		return false, nil
	}
	// return false admin if an error has occurred
	return (err == nil), err
}
