package boxes

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

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
	lastEvent  *events.Event
}

// ComputeBox box according to the received boxID.
// The function retrieves events linked to the boxID using received db connector.
// The function can takes optionally the last event which will be retrieve if nil
func Compute(
	ctx context.Context,
	boxID string,
	exec boil.ContextExecutor,
	identities entrypoints.IdentityIntraprocessInterface,
	lastEvent *events.Event,
) (Box, error) {
	// init the computer system
	computer := &computer{
		exec:       exec,
		identities: identities,
		box:        Box{ID: boxID},
		lastEvent:  lastEvent,
	}
	computer.ePlayer = map[string]func(context.Context, events.Event) error{
		// NOTE: to add an new event here should involve attention on the RequireToBuild method
		// used to retrieve events to compute the box
		"create":          computer.playCreate,
		"state.lifecycle": computer.playState,
	}

	// automatically retrieve events if 0 events loaded
	var err error
	computer.events, err = events.ListForBuild(ctx, computer.exec, computer.box.ID)
	if err != nil {
		return computer.box, err
	}

	// comput the box then return it
	err = computer.do(ctx)
	return computer.box, err
}

func (c *computer) do(ctx context.Context) error {
	// replay events from the last (most recent) to first ()
	// to fill the box informations
	totalCount := len(c.events)
	for i := 0; i < totalCount; i++ {
		e := c.events[totalCount-i-1]
		if err := c.playEvent(ctx, e); err != nil {
			return merror.Transform(err).Describef("playing event %s", e.ID)
		}
	}

	return c.handleLast(ctx)
}

func (c *computer) playEvent(ctx context.Context, e events.Event) error {
	// play the event
	if play, ok := c.ePlayer[e.Type]; ok {
		if err := play(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

// take care of binding the last event in the box
func (c *computer) handleLast(ctx context.Context) error {
	// if the box has been closed and the viewer is not the creator or has no token
	// we force the last event to be the close event and we remove the public key
	acc := ajwt.GetAccesses(ctx)
	if (acc == nil || acc.IdentityID != c.creatorID) && c.closeEvent != nil {
		c.lastEvent = c.closeEvent
		c.box.PublicKey = ""
	} else if c.lastEvent == nil { // retrieve last event if not already there
		last, err := events.GetLast(ctx, c.exec, c.box.ID)
		if err != nil {
			return merror.Transform(err).Describe("getting last event")
		}
		c.lastEvent = &last
	}

	identityMap, err := events.MapSenderIdentities(ctx, []events.Event{*c.lastEvent}, c.identities)
	if err != nil {
		return merror.Transform(err).Describe("retrieving identities for view")
	}

	view, err := events.FormatEvent(*c.lastEvent, identityMap)
	if err != nil {
		return merror.Transform(err).Describe("computing view of last event")
	}
	c.box.LastEvent = view
	return nil
}

func (c *computer) playCreate(ctx context.Context, e events.Event) error {
	c.box.CreatedAt = e.CreatedAt
	CreationContent := events.CreationContent{}
	if err := CreationContent.Unmarshal(e.JSONContent); err != nil {
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
	content := events.StateLifecycleContent{}
	if err := e.JSONContent.Unmarshal(&content); err != nil {
		return err
	}
	c.box.Lifecycle = content.State

	if c.box.Lifecycle == "closed" {
		c.closeEvent = &e
	}

	return nil
}
