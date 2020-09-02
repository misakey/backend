package events

import (
	"context"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/types"
)

type StateLifecycleContent struct {
	State string `json:"state"`
}

func (c *StateLifecycleContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c StateLifecycleContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.State, v.Required, v.In("closed")),
	)
}

func IsClosed(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID string,
) (bool, error) {
	jsonQuery := `{"state": "closed"}`
	_, err := findByTypeContent(ctx, exec, boxID, "state.lifecycle", &jsonQuery)
	if err != nil {
		if merror.HasCode(err, merror.NotFoundCode) {
			return false, nil
		}
		return true, merror.Transform(err).Describe("getting closed lifecycle")
	}
	return true, nil
}

func MustBoxBeOpen(ctx context.Context, exec boil.ContextExecutor, boxID string) error {
	closed, err := IsClosed(ctx, exec, boxID)
	if err != nil {
		return err
	}
	if !closed {
		return nil
	}
	return merror.Conflict().Describe("box is closed").
		Detail("lifecycle", merror.DVConflict)
}

func MustBeAdmin(ctx context.Context, exec boil.ContextExecutor, boxID, senderID string) error {
	// only the creator is an admin so we check the sender ID is the one that has created the box
	createEvent, err := GetCreateEvent(ctx, exec, boxID)
	if err != nil {
		return merror.Transform(err).Describe("getting create event")
	}
	if createEvent.SenderID != senderID {
		return merror.Forbidden().Describe("sender not an admin").
			Detail("sender_id", merror.DVForbidden)
	}
	return nil
}
