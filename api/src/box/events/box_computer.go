package events

import (
	"context"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
)

// Box is a volatile object built based on events linked to its ID
type Box struct {
	ID         string      `json:"id"`
	CreatedAt  time.Time   `json:"server_created_at"`
	OwnerOrgID string      `json:"owner_org_id"`
	DatatagID  null.String `json:"datatag_id"`
	Subject    *SenderView `json:"subject"`
	PublicKey  string      `json:"public_key"`
	Title      string      `json:"title"`
	AccessMode string      `json:"access_mode"`

	// aggregated data
	EventsCount null.Int    `json:"events_count,omitempty"`
	Creator     SenderView  `json:"creator"`
	LastEvent   View        `json:"last_event"`
	BoxSettings *BoxSetting `json:"settings,omitempty"`
}

type computer struct {
	// binding of event type - play func
	ePlayer map[string]func(context.Context, Event) error

	events []Event

	exec       boil.ContextExecutor
	identities *IdentityMapper

	// the output
	box Box

	// local save for internal logic
	creatorID string
	lastEvent *Event
}

// computeBox box according to the received boxID.
// The function retrieves events linked to the boxID using received db connector.
// The function can takes optionally the last event which will be retrieve if nil
func computeBox(
	ctx context.Context,
	boxID string,
	exec boil.ContextExecutor,
	identities *IdentityMapper,
	lastEvent *Event,
) (Box, error) {
	// init the computer system
	computer := &computer{
		exec:       exec,
		identities: identities,
		box:        Box{ID: boxID},
		lastEvent:  lastEvent,
	}
	computer.ePlayer = map[string]func(context.Context, Event) error{
		// NOTE: to add an new event here should involve attention on the RequireToBuild method
		// used to retrieve events to compute the box
		etype.Create:          computer.playCreate,
		etype.Stateaccessmode: computer.playStateAccessMode,
	}

	// automatically retrieve events if 0 events loaded
	var err error
	computer.events, err = ListForBuild(ctx, computer.exec, computer.box.ID)
	if err != nil {
		return computer.box, err
	}

	// compute the box then return it
	err = computer.do(ctx)
	return computer.box, err
}

func (c *computer) do(ctx context.Context) error {
	// 1. before playing all event, set default values:
	c.box.AccessMode = LimitedMode

	// 2. replay events from the last (most recent) to first ()
	// to fill the box informations
	totalCount := len(c.events)
	for i := 0; i < totalCount; i++ {
		e := c.events[totalCount-i-1]
		if err := c.playEvent(ctx, e); err != nil {
			return merr.From(err).Descf("playing event %s", e.ID)
		}
	}

	// handle the last event that have been potentially set during event plays.
	return c.handleLast(ctx)
}

func (c *computer) playEvent(ctx context.Context, e Event) error {
	// owner org id is set in all events so it may be use later for optimization purpose:
	// (aka invalidating caches)
	e.ownerOrgID = null.StringFrom(c.box.OwnerOrgID)

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
	// 1. retrieve last event if not already there
	if c.lastEvent == nil {
		last, err := GetLast(ctx, c.exec, c.box.ID)
		if err != nil {
			return merr.From(err).Desc("getting last event")
		}
		c.lastEvent = &last
	}

	// 2. build event aggregate
	if err := BuildAggregate(ctx, c.exec, c.lastEvent); err != nil {
		return merr.From(err).Desc("building aggregate")
	}

	// 3. format and bind in non-transparent view mode the last event
	view, err := c.lastEvent.Format(ctx, c.identities, false)
	if err != nil {
		return merr.From(err).Desc("computing view of last event")
	}
	c.box.LastEvent = view
	return nil
}

func (c *computer) playCreate(ctx context.Context, e Event) error {
	acc := oidc.GetAccesses(ctx)
	c.box.CreatedAt = e.CreatedAt
	creationContent := CreationContent{}
	if err := creationContent.Unmarshal(e.JSONContent); err != nil {
		return err
	}
	c.box.OwnerOrgID = creationContent.OwnerOrgID
	c.box.DatatagID = null.StringFromPtr(creationContent.DatatagID)
	c.box.PublicKey = creationContent.PublicKey
	c.box.Title = creationContent.Title

	// save the creator id for future logic - data obfuscation
	c.creatorID = e.SenderID

	// set the creator information
	var err error
	// need transparency on creator email if the connected user is the creator
	// so the client can attest either the user is creator or
	// NOTE: must change implementing advanced admin role
	c.box.Creator, err = c.identities.Get(ctx, e.SenderID, (acc != nil && acc.IdentityID == e.SenderID))
	if err != nil {
		return merr.From(err).Desc("retrieving creator")
	}

	// set the subject information if there is any
	if creationContent.SubjectIdentityID != nil {
		subject, err := c.identities.Get(ctx, *creationContent.SubjectIdentityID, (acc != nil && acc.IdentityID == e.SenderID))
		if err != nil {
			return merr.From(err).Desc("retrieving subject")
		}
		c.box.Subject = &subject
	}

	return nil
}

func (c *computer) playStateAccessMode(_ context.Context, e Event) error {
	accessModeContent := AccessModeContent{}
	if err := accessModeContent.Unmarshal(e.JSONContent); err != nil {
		return err
	}
	c.box.AccessMode = accessModeContent.Value
	return nil
}
