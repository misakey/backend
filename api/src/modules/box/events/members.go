package events

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

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

	events, err := ListByBoxIDAndType(ctx, exec, boxID, "join")
	if err != nil {
		return merror.Transform(err).Describe("getting join events")
	}

	for _, event := range events {
		if event.SenderID == senderID {
			return nil
		}
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
