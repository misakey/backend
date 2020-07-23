package boxes

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/eventscounts"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

func Get(ctx context.Context, dbConn *sql.DB, identities entrypoints.IdentityIntraprocessInterface, boxID string) (Box, error) {
	return Compute(ctx, boxID, dbConn, identities)
}

func CountForSender(ctx context.Context, exec boil.Executor, senderID string) (int, error) {
	boxIDs, err := latestIDsForSender(ctx, exec, senderID)
	return len(boxIDs), err
}

func ListForSender(
	ctx context.Context,
	dbConn *sql.DB,
	redConn *redis.Client,
	identities entrypoints.IdentityIntraprocessInterface,
	senderID string,
	limit, offset int,
) ([]*Box, error) {
	boxes := []*Box{}

	// 1. retrieve box IDs
	boxIDs, err := latestIDsForSender(ctx, dbConn, senderID)
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
		box, err := Compute(ctx, boxID, dbConn, identities)
		if err != nil {
			return boxes, merror.Transform(err).Describef("computing box %s", boxID)
		}
		boxes[i] = &box
	}

	// 4. add the new events count for the requesting identity
	eventsCount, err := eventscounts.GetForIdentity(ctx, redConn, senderID)
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

func latestIDsForSender(
	ctx context.Context,
	exec boil.Executor,
	senderID string,
) ([]string, error) {
	bCol := sqlboiler.EventColumns.BoxID
	cCol := sqlboiler.EventColumns.CreatedAt
	sCol := sqlboiler.EventColumns.SenderID
	query := fmt.Sprintf(`
		SELECT %s, max(%s) latest FROM event WHERE %s IN (
			SELECT %s FROM event WHERE %s = '%s'
		) GROUP BY %s ORDER BY latest DESC;
	`, bCol, cCol, bCol, bCol, sCol, senderID, bCol)

	var dbEvents []events.Event
	if err := queries.Raw(query).Bind(ctx, exec, &dbEvents); err != nil {
		return nil, merror.Transform(err).Describe("retrieving box id by sender id")
	}

	// return an empty list if no record was found
	if len(dbEvents) == 0 {
		return []string{}, nil
	}

	boxIDs := make([]string, len(dbEvents))
	for i, record := range dbEvents {
		boxIDs[i] = record.BoxID
	}
	return boxIDs, nil
}
