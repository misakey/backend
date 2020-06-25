package boxes

import (
	"context"
	"time"

	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// Box is a volatile object built based on events linked to its ID
type Box struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"server_created_at"`
	PublicKey string    `json:"public_key"`
	Title     string    `json:"title"`
	Lifecycle string    `json:"lifecycle"`

	// aggregated data
	Creator   events.SenderView `json:"creator"`
	LastEvent events.View       `json:"last_event"`
}

func MustBeOpen(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID string,
) error {
	_, err := events.GetStateLifecycle(ctx, exec, boxID, "closed")
	if err != nil {
		// event closed not found = is not closed
		if merror.HasCode(err, merror.NotFoundCode) {
			return nil
		}
		return merror.Transform(err).Describe("getting lifecycle event")
	}
	return merror.Conflict().Describe("box is closed")
}

func MustBeCreator(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID, senderID string,
) error {
	// first the creator so we can skip
	// checking for the box to be open
	createEvent, err := events.GetCreateEvent(ctx, exec, boxID)
	if err != nil {
		return merror.Transform(err).Describe("getting create event")
	}
	if createEvent.SenderID != senderID {
		return merror.Forbidden().Describe("sender not the creator")
	}
	return nil
}

func MustBeCreatorIfClosed(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID, senderID string,
) error {
	// first the creator so we can skip
	// checking for the box to be open
	createEvent, err := events.GetCreateEvent(ctx, exec, boxID)
	if err != nil {
		return merror.Transform(err).Describe("getting create event")
	}
	if createEvent.SenderID == senderID {
		return nil
	}

	// check if the box is closed
	if _, err := events.GetStateLifecycle(ctx, exec, boxID, "closed"); err != nil {
		// return no err if not found = is not closed
		if merror.HasCode(err, merror.NotFoundCode) {
			return nil
		}
		return merror.Transform(err).Describe("getting lifecycle event")
	}
	return merror.Forbidden().Describe("restricted to creator since box is closed")
}
