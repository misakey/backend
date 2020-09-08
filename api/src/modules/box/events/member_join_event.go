package events

import (
	"context"
	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func joinHandler(ctx context.Context, e *Event, exec boil.ContextExecutor, redConn *redis.Client, identities entrypoints.IdentityIntraprocessInterface) error {
	// check that the current sender is not already a box member
	isMember, err := isMember(ctx, exec, e.BoxID, e.SenderID)
	if err != nil {
		return merror.Transform(err).Describe("checking membership")
	}
	if isMember {
		return merror.Conflict().Describe("already box member")
	}

	// check accesses
	if err := MustHaveAccess(ctx, exec, identities, e.BoxID, e.SenderID); err != nil {
		return merror.Transform(err).Describe("checking accesses")
	}

	return nil
}

// List box ids joined by an identity ID
func ListJoinedBoxIDs(ctx context.Context, exec boil.ContextExecutor, senderID string) ([]string, error) {
	joinEvents, err := list(ctx, exec, eventFilters{
		eType:     null.StringFrom("member.join"),
		unrefered: true,
		senderID:  null.StringFrom(senderID),
	})
	if err != nil {
		return nil, err
	}

	boxIDs := make([]string, len(joinEvents))
	for i, e := range joinEvents {
		boxIDs[i] = e.BoxID
	}
	return boxIDs, nil
}
