package events

import (
	"context"

	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type JoinContent struct{}

func (c *JoinContent) Unmarshal(json types.JSON) error {
	return json.Unmarshal(c)
}

func (c JoinContent) Validate() error {
	return nil
}

func StoreJoin(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) error {

	// check that the current sender has not already
	// joined this box
	_, err := ListByTypeAndBoxIDAndSenderID(ctx, exec, "join", boxID, senderID)
	if err == nil {
		// if there is no not found error
		// then a join already exist for the current sender
		// and we don’t need to add a new one
		return nil
	}
	if !merror.HasCode(err, merror.NotFoundCode) {
		return err
	}

	// create and store the new join event
	event, err := newWithAnyContent("join", nil, boxID, senderID)
	if err != nil {
		return err
	}

	if err := event.ToSQLBoiler().Insert(ctx, exec, boil.Infer()); err != nil {
		return merror.Transform(err).Describe("inserting event in DB")
	}

	return nil
}
