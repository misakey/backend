package events

import (
	"context"

	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

func StoreJoin(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) error {

	// check that the current sender is not already a box member
	if err := MustBeMember(ctx, exec, boxID, senderID); err == nil {
		// user is a box member
		// so we just return
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
