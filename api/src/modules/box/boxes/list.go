package boxes

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func Get(ctx context.Context, exec boil.ContextExecutor, identities entrypoints.IdentityIntraprocessInterface, boxID string) (Box, error) {
	return Compute(ctx, boxID, exec, identities, nil)
}

func CountForSender(ctx context.Context, exec boil.ContextExecutor, senderID string) (int, error) {
	boxIDs, err := ListSenderBoxIDs(ctx, exec, senderID)
	return len(boxIDs), err
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
	// 1. retrieve box IDs
	boxIDs, err := ListSenderBoxIDs(ctx, exec, senderID)
	if err != nil {
		return boxes, merror.Transform(err).Describe("listing box ids")
	}

	// 2. order by most recent and put pagination in place
	// TODO (perf): this query does not use any index and is quite heavy
	mods := []qm.QueryMod{
		qm.Select("box_id", "max(created_at)"),
		sqlboiler.EventWhere.BoxID.IN(boxIDs),
		qm.GroupBy("box_id"),
		qm.OrderBy("max DESC"),
		qm.Offset(offset),
		qm.Limit(limit),
	}

	lastEvents, err := sqlboiler.Events(mods...).All(ctx, exec)
	if err != nil {
		return boxes, merror.Transform(err).Describe("ordering boxes")
	}

	// 3. compute all boxes
	boxes = make([]*Box, len(lastEvents))
	for i, e := range lastEvents {
		// TODO (perf): computation in redis
		box, err := Compute(ctx, e.BoxID, exec, identities, nil)
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

func ListSenderBoxIDs(
	ctx context.Context,
	exec boil.ContextExecutor,
	senderID string,
) ([]string, error) {
	joinedBoxIDs, err := events.ListMemberBoxIDs(ctx, exec, senderID)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing joined box ids")
	}
	createdBoxIDs, err := events.ListCreatorBoxIDs(ctx, exec, senderID)
	if err != nil {
		return nil, merror.Transform(err).Describe("listing creator box ids")
	}

	ids := append(joinedBoxIDs, createdBoxIDs...)
	var uniqueIDs []string
	addedOnes := make(map[string]bool)
	for _, boxID := range ids {
		_, ok := addedOnes[boxID]
		if !ok {
			uniqueIDs = append(uniqueIDs, boxID)
			addedOnes[boxID] = true
		}
	}
	return uniqueIDs, nil
}
