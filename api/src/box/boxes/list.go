package boxes

import (
	"context"
	"sort"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/slice"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events/cache"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/box/repositories/sqlboiler"
)

// Get ...
func Get(ctx context.Context, exec boil.ContextExecutor, identities *events.IdentityMapper, boxID string) (events.Box, error) {
	return events.Compute(ctx, boxID, exec, identities, nil)
}

// GetWithSenderInfo ...
func GetWithSenderInfo(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, identities *events.IdentityMapper, boxID, identityID string) (*events.Box, error) {
	box, err := events.Compute(ctx, boxID, exec, identities, nil)
	if err != nil {
		return nil, err
	}

	// fill the eventCounts attribute
	eventsCount, err := events.GetCountForIdentity(ctx, redConn, identityID, boxID)
	if err != nil {
		return nil, merr.From(err).Desc("counting new events")
	}
	box.EventsCount = null.IntFrom(eventsCount)

	boxSetting, err := events.GetBoxSetting(ctx, exec, identityID, boxID)
	if err != nil {
		return nil, merr.From(err).Desc("getting box setting")
	}
	box.BoxSettings = boxSetting

	return &box, nil
}

// CountForSender ...
func CountForSender(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, senderID string) (int, error) {
	list, err := LastSenderBoxEvents(ctx, exec, redConn, senderID, []string{})
	return len(list), err
}

// ListSenderBoxes ...
func ListSenderBoxes(
	ctx context.Context,
	exec boil.ContextExecutor,
	redConn *redis.Client,
	identities *events.IdentityMapper,
	senderID string,
	limit, offset int,
) ([]*events.Box, error) {
	boxes := []*events.Box{}
	// 1. retrieve lastest events concerning the user's boxes
	list, err := LastSenderBoxEvents(ctx, exec, redConn, senderID, etype.MembersCanSee)
	if err != nil {
		return boxes, merr.From(err).Desc("listing box ids")
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
	boxIDs := make([]string, len(list))
	boxes = make([]*events.Box, len(list))
	for i, e := range list {
		// TODO (perf): computation in redis
		box, err := events.Compute(ctx, e.BoxID, exec, identities, &e)
		if err != nil {
			return boxes, merr.From(err).Descf("computing box %s", e.BoxID)
		}
		boxes[i] = &box
		boxIDs[i] = box.ID
	}

	// 4. retrieve box settings
	settingsFilters := events.BoxSettingFilters{
		BoxIDs:     boxIDs,
		IdentityID: senderID,
	}
	boxSettings, err := events.ListBoxSettings(ctx, exec, settingsFilters)
	if err != nil {
		return boxes, merr.From(err).Desc("listing box settings")
	}
	indexedBoxSettings := make(map[string]events.BoxSetting, len(boxSettings))
	for _, boxSetting := range boxSettings {
		indexedBoxSettings[boxSetting.BoxID] = *boxSetting
	}

	// 5. add events count and box settings data to boxes
	for _, box := range boxes {
		// we won’t return an error since the list
		// can still be returned
		// with a wrong amount of event counts
		box.EventsCount = null.IntFrom(events.ComputeCount(ctx, redConn, senderID, box.ID))

		// add box settings
		boxSetting, ok := indexedBoxSettings[box.ID]
		if !ok {
			boxSetting = *events.GetDefaultBoxSetting(senderID, box.ID)
		}
		box.BoxSettings = &boxSetting
	}

	return boxes, nil
}

// LastSenderBoxIDs ...
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
		return nil, merr.From(err).Desc("listing joined box ids")
	}
	creates, err := events.ListCreatorIDEvents(ctx, exec, senderID)
	if err != nil {
		return nil, merr.From(err).Desc("listing creator box ids")
	}

	// it is forbidden to join box the user has created so we already have unique box IDs
	boxIDs := make([]string, len(joins)+len(creates))
	idx := 0
	for _, event := range append(joins, creates...) {
		boxIDs[idx] = event.BoxID
		idx++
	}

	// 3. update cache
	if len(boxIDs) > 0 {
		if _, err := redConn.SAdd(cache.GetSenderBoxesKey(senderID), slice.StringSliceToInterfaceSlice(boxIDs)...).Result(); err != nil {
			logger.FromCtx(ctx).Warn().Err(err).Msgf("could not build boxes cache for %s", senderID)
		}
	}

	return boxIDs, nil
}

// LastSenderBoxEvents ...
func LastSenderBoxEvents(
	ctx context.Context,
	exec boil.ContextExecutor,
	redConn *redis.Client,
	senderID string,
	etypes []string,
) ([]events.Event, error) {
	boxIDs, err := LastSenderBoxIDs(ctx, exec, redConn, senderID)
	if err != nil {
		return nil, err
	}

	mods := []qm.QueryMod{
		qm.Select("DISTINCT ON (box_id) box_id, event.*"),
		sqlboiler.EventWhere.BoxID.IN(boxIDs),
		qm.OrderBy(sqlboiler.EventColumns.BoxID),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt + " DESC"),
	}

	if len(etypes) > 0 {
		mods = append(mods, sqlboiler.EventWhere.Type.IN(etypes))
	}

	lastEventsDB, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return []events.Event{}, merr.From(err).Desc("retrieving last events")
	}

	// get last events
	lastEvents := make([]events.Event, len(lastEventsDB))
	idx := 0
	for _, event := range lastEventsDB {
		lastEvents[idx] = events.FromSQLBoiler(event)
		idx++
	}

	sort.Slice(lastEvents, func(i, j int) bool { return lastEvents[i].CreatedAt.Unix() > lastEvents[j].CreatedAt.Unix() })
	return lastEvents, nil
}
