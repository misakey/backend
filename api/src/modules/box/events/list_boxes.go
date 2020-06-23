package events

import (
	"context"
	"fmt"

	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

func CountSenderBoxes(ctx context.Context, exec boil.ContextExecutor, senderID string) (int, error) {
	boxIDs, err := lastestBoxIDsForSender(ctx, exec, senderID)
	return len(boxIDs), err
}

func GetSenderBoxes(
	ctx context.Context,
	repo repositories.Contextual,
	senderID string,
	limit int,
	offset int,
) ([]Box, error) {
	boxes := []Box{}

	// 1. retrieve box IDs
	boxIDs, err := lastestBoxIDsForSender(ctx, repo.DB(), senderID)
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
	boxes = make([]Box, len(boxIDs))
	for i, boxID := range boxIDs {
		boxes[i], err = ComputeBox(ctx, boxID, repo)
		if err != nil {
			return boxes, merror.Transform(err).Describef("computing box %s", boxID)
		}
	}
	return boxes, nil
}

func lastestBoxIDsForSender(
	ctx context.Context,
	exec boil.ContextExecutor,
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

	var dbEvents []Event
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
