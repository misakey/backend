package boxes

import (
	"context"
	"sort"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/slice"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func Get(ctx context.Context, exec boil.ContextExecutor, identities entrypoints.IdentityIntraprocessInterface, boxID string) (Box, error) {
	return Compute(ctx, boxID, exec, identities, nil)
}

func CountForSender(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, senderID string) (int, error) {
	list, err := LastSenderBoxEvents(ctx, exec, redConn, senderID)
	return len(list), err
}

func ListSenderBoxes(
	ctx context.Context,
	exec boil.ContextExecutor,
	redConn *redis.Client,
	identities entrypoints.IdentityIntraprocessInterface,
	senderID string,
	limit, offset int,
) ([]*Box, error) {
	boxes := []*Box{}
	// 1. retrieve lastest events concerning the user's boxes
	list, err := LastSenderBoxEvents(ctx, exec, redConn, senderID)
	if err != nil {
		return boxes, merror.Transform(err).Describe("listing box ids")
	}

	// 2. put pagination in place
	// if the offset is higher than the total size, we return an empty list
	if offset >= len(list) {
		return boxes, nil
	}
	// cut the slice using the offset
	list = list[offset:]
	// cut the slice using the limit
	if len(list) > limit {
		list = list[:limit]
	}

	// 3. compute all boxes
	boxes = make([]*Box, len(list))
	for i, e := range list {
		// TODO (perf): computation in redis
		box, err := Compute(ctx, e.BoxID, exec, identities, &e)
		if err != nil {
			return boxes, merror.Transform(err).Describef("computing box %s", e.BoxID)
		}
		boxes[i] = &box
	}

	// 4. add the new events count for the requesting identity
	eventsCount, err := events.GetCountsForIdentity(ctx, redConn, senderID)
	if err != nil {
		return nil, merror.Transform(err).Describe("getting new events count")
	}
	for _, box := range boxes {
		// if there is no value for a given box
		// that means no new event since last visit
		count, ok := eventsCount[box.ID]
		if !ok {
			box.EventsCount = 0
		}
		box.EventsCount = count
	}

	// 5. eventually return the boxes list
	return boxes, nil
}

func LastSenderBoxIDs(
	ctx context.Context,
	exec boil.ContextExecutor,
	redConn *redis.Client,
	senderID string,
) ([]string, error) {
	// 1. try to retrieve cache
	cacheBoxIDs, err := redConn.SMembers(cache.GetSenderBoxesKey(senderID)).Result()
	if err == nil && len(cacheBoxIDs) != 0 {
		return cacheBoxIDs, nil
	}

	// 2. build list
	joins, err := events.ListMemberBoxLatestEvents(ctx, exec, senderID)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing joined box ids")
	}
	creates, err := events.ListCreatorIDEvents(ctx, exec, senderID)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing creator box ids")
	}

	// it is forbidden to join box the user has created so we already have unique box IDs
	boxIDs := make([]string, len(joins)+len(creates))
	idx := 0
	for _, event := range append(joins, creates...) {
		boxIDs[idx] = event.BoxID
		idx += 1
	}

	// 3. update cache
	if _, err := redConn.SAdd(cache.GetSenderBoxesKey(senderID), slice.StringSliceToInterfaceSlice(boxIDs)...).Result(); err != nil {
		logger.FromCtx(ctx).Warn().Err(err).Msgf("could not build boxes cache for %s", senderID)
	}

	return boxIDs, nil
}

func LastSenderBoxEvents(
	ctx context.Context,
	exec boil.ContextExecutor,
	redConn *redis.Client,
	senderID string,
) ([]events.Event, error) {
	boxIDs, err := LastSenderBoxIDs(ctx, exec, redConn, senderID)
	if err != nil {
		return nil, err
	}

	mods := []qm.QueryMod{
		qm.Select("DISTINCT ON (box_id) box_id, event.*"),
		sqlboiler.EventWhere.BoxID.IN(boxIDs),
		qm.OrderBy("box_id"),
		qm.OrderBy("created_at DESC"),
	}

	lastEventsDB, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return []events.Event{}, merror.Transform(err).Describe("retrieving last events")
	}

	// get last events
	lastEvents := make([]events.Event, len(lastEventsDB))
	idx := 0
	for _, event := range lastEventsDB {
		lastEvents[idx] = events.FromSQLBoiler(event)
		idx += 1
	}

	sort.Slice(lastEvents, func(i, j int) bool { return lastEvents[i].CreatedAt.Unix() > lastEvents[j].CreatedAt.Unix() })
	return lastEvents, nil
}
