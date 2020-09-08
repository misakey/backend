package boxes

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func Get(ctx context.Context, exec boil.ContextExecutor, identities entrypoints.IdentityIntraprocessInterface, boxID string) (Box, error) {
	return Compute(ctx, boxID, exec, identities)
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

	// 2. put pagination in place
	// if the offset is higher than the total size, we return an empty list
	if offset >= len(boxIDs) {
		return boxes, nil
	}
	// cut the slice using the offset
	boxIDs = boxIDs[offset:]
	// cut the slice using the limit
	if len(boxIDs) > limit {
		boxIDs = boxIDs[:limit]
	}

	// 3. compute all boxes
	boxes = make([]*Box, len(boxIDs))
	for i, boxID := range boxIDs {
		box, err := Compute(ctx, boxID, exec, identities)
		if err != nil {
			return boxes, merror.Transform(err).Describef("computing box %s", boxID)
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
	joinedBoxIDs, err := events.ListJoinedBoxIDs(ctx, exec, senderID)
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
