package events

import (
	"context"
	"database/sql"
	"time"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
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
	Creator   SenderView `json:"creator"`
	LastEvent View       `json:"last_event"`
}

type boxComputer struct {
	// binding of event type - play func
	ePlayer map[string]func(context.Context, Event) error

	// repositories
	identityRepo entrypoints.IdentityIntraprocessInterface
	db           *sql.DB

	events []Event

	// the output
	box *Box
}

// ComputeBox box according to the received boxID.
// The function retrieves events linked to the boxID using received repository actors.
// The function can takes optionally events, only these events will be used to compute
// the box if there are some.
func ComputeBox(
	ctx context.Context,
	boxID string,
	db *sql.DB, identityRepo entrypoints.IdentityIntraprocessInterface,
	events ...Event,
) (*Box, error) {
	// init the computer system
	computer := &boxComputer{
		db:           db,
		identityRepo: identityRepo,
		box:          &Box{ID: boxID},
		events:       events,
	}
	computer.ePlayer = map[string]func(context.Context, Event) error{
		"create":          computer.playCreate,
		"state.lifecycle": computer.playState,
	}

	// automatically retrieve events if 0 events loaded
	if len(events) == 0 {
		computer.retrieveEvents(ctx)
	}

	// comput the box then return it
	err := computer.do(ctx)
	return computer.box, err
}

func (c *boxComputer) retrieveEvents(ctx context.Context) error {
	var err error
	c.events, err = List(ctx, c.box.ID, c.db)
	return err
}

func (c *boxComputer) do(ctx context.Context) error {
	// replay events from the last (most recent) to first ()
	// to fill the box informations
	totalCount := len(c.events)
	for i := 0; i < totalCount; i++ {
		e := c.events[totalCount-i-1]
		if err := c.playEvent(ctx, e, i == totalCount-1); err != nil {
			return merror.Transform(err).Describef("playing event %s", e.ID)
		}
	}
	return nil
}

func (c *boxComputer) playEvent(ctx context.Context, e Event, last bool) error {
	// play the event
	if play, ok := c.ePlayer[e.Type]; ok {
		if err := play(ctx, e); err != nil {
			return err
		}
	}

	// take care of binding the last event in the box
	if last {
		// get the sender information
		identity, err := c.identityRepo.GetIdentity(ctx, e.SenderID)
		if err != nil {
			return merror.Transform(err).Describe("retrieving last sender")
		}
		c.box.LastEvent = ToView(e, identity)
	}
	return nil
}
