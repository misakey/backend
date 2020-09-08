package events

import (
	"context"
	"github.com/volatiletech/null"
	"github.com/volatiletech/sqlboiler/boil"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
)

func StoreJoin(
	ctx context.Context, exec boil.ContextExecutor, identities entrypoints.IdentityIntraprocessInterface,
	boxID, senderID string,
) error {
	// check that the current sender is not already a box member
	isMember, err := isMember(ctx, exec, boxID, senderID)
	if err != nil {
		return err
	}
	// user is already a member so we just return
	if isMember {
		return nil
	}

	// create and store the new join event
	event, err := newWithAnyContent("member.join", nil, boxID, senderID)
	if err != nil {
		return err
	}

	if err := event.ToSQLBoiler().Insert(ctx, exec, boil.Infer()); err != nil {
		return merror.Transform(err).Describe("inserting event in DB")
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
