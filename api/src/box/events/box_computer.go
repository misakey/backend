package events

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
)

// Box is a volatile object built based on events linked to its ID
type Box struct {
	ID          string      `json:"id"`
	CreatedAt   time.Time   `json:"server_created_at"`
	OwnerOrgID  string      `json:"owner_org_id"`
	DatatagID   null.String `json:"datatag_id"`
	DataSubject *string     `json:"data_subject,omitempty"`
	PublicKey   string      `json:"public_key"`
	Title       string      `json:"title"`
	AccessMode  string      `json:"access_mode"`

	// internal computation logic
	creatorID string
}

type computer struct {
	// binding of event type - play func
	ePlayer map[string]func(context.Context, Event) error

	events []Event

	exec           boil.ContextExecutor
	identityMapper *IdentityMapper

	// the output
	box Box
}

// computeBox box according to the received boxID.
// The function retrieves events linked to the boxID using received db connector.
func computeBox(
	ctx context.Context,
	boxID string,
	exec boil.ContextExecutor,
	identityMapper *IdentityMapper,
) (Box, error) {
	// init the computer system
	computer := &computer{
		exec:           exec,
		identityMapper: identityMapper,
		box:            Box{ID: boxID},
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

type BoxView struct {
	Box

	// aggregated data
	Subject     *SenderView `json:"subject"`
	Creator     SenderView  `json:"creator"`
	LastEvent   View        `json:"last_event"`
	EventsCount null.Int    `json:"events_count,omitempty"`
	BoxSettings *BoxSetting `json:"settings,omitempty"`

	lastEvent *Event
}

// SetLastEvent which avoid its retrieval in case of nil
func SetLastEvent(e *Event) func(*BoxView) {
	return func(boxView *BoxView) {
		boxView.lastEvent = e
	}
}

// computeBoxView according to the received boxID.
// use computeBox to assembly a box then add some additionnaly information
// used for box view purpose
func computeBoxView(
	ctx context.Context,
	exec boil.ContextExecutor, identityMapper *IdentityMapper, redConn *redis.Client,
	boxID string,
	options ...func(*BoxView),
) (view BoxView, err error) {
	box, err := computeBox(ctx, boxID, exec, identityMapper)
	if err != nil {
		return view, err
	}
	view.Box = box

	// run options
	for _, option := range options {
		option(&view)
	}

	// creator view
	// NOTE: need transparency on creator email if the connected user is an admin
	// NOTE: must change implementing advanced admin role
	acc := oidc.GetAccesses(ctx)
	isAdmin := (acc != nil && acc.IdentityID == box.creatorID)
	view.Creator, err = identityMapper.Get(ctx, box.creatorID, isAdmin)
	if err != nil {
		return view, merr.From(err).Desc("retrieving creator")
	}

	// dDta Subject view
	if box.DataSubject != nil {
		// the subject identifier must be always shared despite being an admin or not
		dataSubject, err := identityMapper.GetByIdentifierValue(ctx, *box.DataSubject)
		if err != nil {
			return view, merr.From(err).Desc("retrieving subject")
		}
		view.Subject = &dataSubject
	}

	// Last Event view
	// retrieve last event if not already there
	if view.lastEvent == nil {
		last, err := GetLast(ctx, exec, boxID)
		if err != nil {
			return view, merr.From(err).Desc("getting last event")
		}
		view.lastEvent = &last
	}
	// build event aggregate
	if err := BuildAggregate(ctx, exec, view.lastEvent); err != nil {
		return view, merr.From(err).Desc("building aggregate")
	}
	// format and bind in non-transparent view mode the last event
	formattedEvent, err := view.lastEvent.Format(ctx, identityMapper, false)
	if err != nil {
		return view, merr.From(err).Desc("computing view of last event")
	}
	view.LastEvent = formattedEvent

	if acc != nil {
		// bind count event for the current identity
		eventsCount, err := CountEventsBoxForIdentity(ctx, redConn, acc.IdentityID, boxID)
		if err != nil {
			return view, merr.From(err).Desc("counting events for identity")
		}
		view.EventsCount = null.IntFrom(eventsCount)

		// bind box settings for the current identity
		view.BoxSettings, err = GetBoxSettings(ctx, exec, acc.IdentityID, boxID)
		if err != nil {
			if !merr.IsANotFound(err) { // do not return error on not found
				return view, merr.From(err).Desc("getting box setting")
			}
			// but set a default value
			view.BoxSettings = GetDefaultBoxSetting(acc.IdentityID, boxID)
		}
	}
	return view, nil
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
	return nil
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

func (c *computer) playCreate(ctx context.Context, e Event) error {
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
	c.box.creatorID = e.SenderID

	// compute data subject identifier value
	if creationContent.SubjectIdentityID != nil {
		// the subject identifier must be always shared despite being an admin or not
		dataSubject, err := c.identityMapper.Get(ctx, *creationContent.SubjectIdentityID, true)
		if err != nil {
			return merr.From(err).Desc("retrieving subject")
		}
		c.box.DataSubject = &dataSubject.IdentifierValue
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
