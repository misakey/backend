package boxes

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

// Box is a volatile object built based on events linked to its ID
type Box struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"server_created_at"`
	PublicKey string    `json:"public_key"`
	Title     string    `json:"title"`
	Lifecycle string    `json:"lifecycle"`

	// aggregated data
	EventsCount int               `json:"events_count"`
	Creator     events.SenderView `json:"creator"`
	LastEvent   events.View       `json:"last_event"`
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
	return merror.Conflict().Describe("box is closed").
		Detail("lifecycle", merror.DVConflict)
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
		return merror.Forbidden().Describe("sender not the creator").
			Detail("sender_id", merror.DVForbidden)
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
	return merror.Forbidden().Describe("restricted to creator since box is closed").
		Detail("sender_id", merror.DVForbidden)
}

func MustBeActor(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID, senderID string,
) error {
	// if creator, returns immediatly
	if err := MustBeCreator(ctx, exec, boxID, senderID); err == nil {
		return err
	}

	events, err := events.ListByBoxIDAndType(ctx, exec, boxID, "join")
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

func GetActorsExcept(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID, exceptID string,
) ([]string, error) {
	// we build a set to find all uniq actors
	uniqActors := make(map[string]bool)
	events, err := events.ListByBoxID(ctx, exec, boxID)
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

func CheckBoxExists(ctx context.Context, boxID string, exec boil.ContextExecutor) (bool, error) {
	_, err := sqlboiler.Events(
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		sqlboiler.EventWhere.Type.EQ("create"),
	).One(ctx, exec)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		} else {
			return false, merror.Transform(err).Describe("retrieving box creation event")
		}
	}
	return true, nil
}
