package boxes

import (
	"context"

	"github.com/volatiletech/sqlboiler/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/identities"
)

type computer struct {
	// binding of event type - play func
	ePlayer map[string]func(context.Context, events.Event) error

	events []events.Event

	exec       boil.ContextExecutor
	identities entrypoints.IdentityIntraprocessInterface

	// the output
	box Box

	// local save for internal logic
	creatorID  string
	closeEvent *events.Event
}

// ComputeBox box according to the received boxID.
// The function retrieves events linked to the boxID using received db connector.
// The function can takes optionally events, only these events will be used to compute
// the box if there are some.
func Compute(
	ctx context.Context,
	boxID string,
	exec boil.ContextExecutor,
	identities entrypoints.IdentityIntraprocessInterface,
	buildEvents ...events.Event,
) (Box, error) {
	// init the computer system
	computer := &computer{
		exec:       exec,
		identities: identities,
		box:        Box{ID: boxID},
		events:     buildEvents,
	}
	computer.ePlayer = map[string]func(context.Context, events.Event) error{
		"create":          computer.playCreate,
		"state.lifecycle": computer.playState,
	}

	// automatically retrieve events if 0 events loaded
	if len(buildEvents) == 0 {
		if err := computer.retrieveEvents(ctx); err != nil {
			return computer.box, err
		}
	}

	// comput the box then return it
	err := computer.do(ctx)
	return computer.box, err
}

func (c *computer) retrieveEvents(ctx context.Context) error {
	var err error
	c.events, err = events.ListByBoxID(ctx, c.exec, c.box.ID)
	return err
}

func (c *computer) do(ctx context.Context) error {
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

func (c *computer) playEvent(ctx context.Context, e events.Event, last bool) error {
	// play the event
	if play, ok := c.ePlayer[e.Type]; ok {
		if err := play(ctx, e); err != nil {
			return err
		}
	}

	// take care of binding the last event in the box
	if last {
		acc := ajwt.GetAccesses(ctx)

		// if the box has been closed and the viewer is not the creator or has no token
		// we force the last event to be the close event
		if (acc == nil || acc.IdentityID != c.creatorID) && c.closeEvent != nil {
			e = *c.closeEvent
			c.box.PublicKey = ""
		}

		identityMap, err := events.MapSenderIdentities(ctx, []events.Event{e}, c.identities)
		if err != nil {
			return merror.Transform(err).Describe("retrieving identities for view")
		}

		view, err := events.ToView(e, identityMap)
		if err != nil {
			return merror.Transform(err).Describe("computing view of last event")
		}

		c.box.LastEvent = view

	}
	return nil
}

func (c *computer) playCreate(ctx context.Context, e events.Event) error {
	c.box.CreatedAt = e.CreatedAt
	CreationContent := events.CreationContent{}
	if err := CreationContent.Unmarshal(e.Content); err != nil {
		return err
	}
	c.box.Lifecycle = "open"
	c.box.PublicKey = CreationContent.PublicKey
	c.box.Title = CreationContent.Title

	// save the creator id for future logic - data obfuscation
	c.creatorID = e.SenderID

	// set the creator information
	identity, err := identities.Get(ctx, c.identities, e.SenderID)
	if err != nil {
		return merror.Transform(err).Describe("retrieving creator")
	}
	c.box.Creator = events.NewSenderView(identity)
	return nil
}

// today, state if only about lifecycle
func (c *computer) playState(_ context.Context, e events.Event) error {
	lifecycleContent := events.StateLifecycleContent{}
	if err := lifecycleContent.Unmarshal(e.Content); err != nil {
		return err
	}
	c.box.Lifecycle = lifecycleContent.State

	if c.box.Lifecycle == "closed" {
		c.closeEvent = &e
	}

	return nil
}
