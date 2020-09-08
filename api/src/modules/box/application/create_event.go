package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

type CreateEventRequest struct {
	// required
	boxID   string
	Type    string     `json:"type"`
	Content types.JSON `json:"content"`

	// optional
	ReferrerID *string `json:"referrer_id"`
}

func (req *CreateEventRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.Type, v.Required, v.In("msg.text", "state.lifecycle", "msg.delete", "msg.edit", "access.add", "access.rm", "member.leave", "member.join")),
		v.Field(&req.ReferrerID, is.UUIDv4),
		v.Field(&req.Content, v.When(events.ContentIsRequired(req.Type), v.Required)),
	)
}

func (bs *BoxApplication) CreateEvent(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*CreateEventRequest)
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	event, err := events.New(req.Type, req.Content, req.boxID, acc.IdentityID, req.ReferrerID)
	if err != nil {
		return nil, err
	}

	view := events.View{}

	// check the box exists and is not closed
	if err := events.MustBoxBeOpen(ctx, bs.db, event.BoxID); err != nil {
		return view, merror.Transform(err).Describe("checking open")
	}

	// call the proper event handler
	if err := events.Handler(event.Type)(ctx, &event, bs.db, bs.redConn, bs.identities); err != nil {
		return view, merror.Transform(err).Describef("handling %s event", event.Type)
	}

	// TODO (code structure): use handlers
	if event.Type == "msg.delete" {
		return bs.deleteMessage(ctx, event)
	}
	// TODO (code structure): use handlers
	if event.Type == "msg.edit" {
		return bs.editMessage(ctx, event)
	}

	if err := event.ToSQLBoiler().Insert(ctx, bs.db, boil.Infer()); err != nil {
		return view, merror.Transform(err).Describe("inserting event in DB")
	}

	// increment count for all identities except the sender
	// only if the event concerns them - today we consider not wished to be notified:
	// - state.lifecycle events
	if event.Type != "state.lifecycle" {
		if err := events.NotifyMembers(ctx, bs.db, bs.redConn, event.SenderID, event.BoxID); err != nil {
			// we log the error but we don’t return it
			logger.FromCtx(ctx).Warn().Err(err).Msg("could not notify members")
		}
	}

	identityMap, err := events.MapSenderIdentities(ctx, []events.Event{event}, bs.identities)
	if err != nil {
		return view, merror.Transform(err).Describe("retrieving identities for view")
	}

	view, err = events.ToView(event, identityMap)
	if err != nil {
		return view, merror.Transform(err).Describe("computing event view")
	}

	return view, nil
}
