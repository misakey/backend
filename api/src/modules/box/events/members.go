package events

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

func ListBoxMembers(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID string,
) ([]string, error) {
	sCol := sqlboiler.EventColumns.SenderID
	rCol := sqlboiler.EventColumns.RefererID
	tCol := sqlboiler.EventColumns.Type
	bCol := sqlboiler.EventColumns.BoxID

	query := fmt.Sprintf(`
		SELECT %s FROM event
		WHERE %s = '%s'
		AND type = 'create'
		UNION
		SELECT %s FROM event
		WHERE %s = '%s'
		AND %s = '%s'
		AND id NOT IN (SELECT %s FROM event WHERE %s = '%s' AND %s = '%s')
		GROUP BY %s;
	`, sCol, bCol, boxID,
		sCol, tCol, "member.join", bCol, boxID, rCol, tCol, "member.leave", bCol, boxID, sCol)

	var dbEvents []Event
	if err := queries.Raw(query).Bind(ctx, exec, &dbEvents); err != nil {
		return nil, merror.Transform(err).Describe("retrieving box members")
	}

	senderIDs := make([]string, len(dbEvents))
	i := 0
	for _, e := range dbEvents {
		senderIDs[i] = e.SenderID
		i += 1
	}

	return senderIDs, nil
}

func getMemberIDsExcept(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID, exceptID string,
) ([]string, error) {
	// we build a set to find all uniq actors
	uniqActors := make(map[string]bool)
	events, err := ListByBoxID(ctx, exec, boxID, nil, nil)
	if err != nil {
		return nil, err
	}
	for _, event := range events {
		if event.SenderID != exceptID {
			uniqActors[event.SenderID] = true
		}
	}

	// we return the list
	actors := make([]string, len(uniqActors))
	idx := 0
	for actor := range uniqActors {
		actors[idx] = actor
		idx += 1
	}

	return actors, nil
}

func MustBeMember(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID, senderID string,
) error {
	// if creator, returns immediatly (performance purpose)
	if err := MustBeAdmin(ctx, exec, boxID, senderID); err == nil {
		return err
	}

	sCol := sqlboiler.EventColumns.SenderID
	rCol := sqlboiler.EventColumns.RefererID
	tCol := sqlboiler.EventColumns.Type
	bCol := sqlboiler.EventColumns.BoxID

	query := fmt.Sprintf(`
			SELECT %s FROM event
			WHERE %s = '%s'
			AND %s = '%s'
			AND %s = '%s'
			AND id NOT IN (SELECT %s FROM event WHERE %s = '%s' AND %s = '%s');
	`, sCol, bCol, boxID, sCol, senderID, tCol, "member.join",
		rCol, tCol, "member.leave", bCol, boxID)

	var dbEvents []Event
	if err := queries.Raw(query).Bind(ctx, exec, &dbEvents); err != nil {
		return merror.Transform(err).Describe("retrieving box id by sender id")
	}

	if len(dbEvents) > 0 {
		return nil
	}

	return merror.Forbidden().Describe("restricted to actor").Detail("sender_id", merror.DVForbidden)
}

func NotifyMembers(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, senderID, boxID string) error {
	// retrieve all member ids excepted the sender id
	memberIDs, err := getMemberIDsExcept(ctx, exec, boxID, senderID)
	if err != nil {
		return merror.Transform(err).Describe("fetching list of members")
	}

	// incr counts for a given box for all received identityIDs
	return incrCounts(ctx, redConn, memberIDs, boxID)
}
