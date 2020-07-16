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

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/eventscounts"
)

type CreateEventRequest struct {
	boxID string

	Type    string     `json:"type"`
	Content types.JSON `json:"content"`
}

func (req *CreateEventRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.Type, v.Required, v.In("msg.text", "state.lifecycle")),
		v.Field(&req.Content, v.Required),
	)
}

func (bs *BoxApplication) CreateEvent(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*CreateEventRequest)
	acc := ajwt.GetAccesses(ctx)

	event, err := events.New(req.Type, req.Content, req.boxID, acc.Subject)
	if err != nil {
		return nil, merror.Transform(err).From(merror.OriBody)
	}

	view := events.View{}

	// check the box does exist
	boxExists, err := boxes.CheckBoxExists(ctx, event.BoxID, bs.db)
	if err != nil {
		return view, merror.Transform(err).Describe("checking existence of box")
	}
	if !boxExists {
		return view, merror.NotFound().Detail("id", merror.DVNotFound).
			Describef("no box with id %s", event.BoxID)
	}

	// check the box is not closed
	if err := boxes.MustBeOpen(ctx, bs.db, event.BoxID); err != nil {
		return view, merror.Transform(err).Describe("checking open")
	}

	// check the sender is creator of the box for state.lifecycle events
	if event.Type == "state.lifecycle" {
		if err := boxes.MustBeCreator(ctx, bs.db, event.BoxID, event.SenderID); err != nil {
			return view, merror.Transform(err).Describe("checking creator")
		}
	}

	sender, err := bs.identities.Get(ctx, event.SenderID)
	if err != nil {
		return view, merror.Transform(err).Describe("fetching sender identity")
	}

	if err := event.ToSQLBoiler().Insert(ctx, bs.db, boil.Infer()); err != nil {
		return view, merror.Transform(err).Describe("inserting event in DB")
	}

	// increment count for all identities except the sender
	// only if this is not a lifecycle event
	if event.Type != "state.lifecycle" {
		identities, err := boxes.GetActorsExcept(ctx, bs.db, event.BoxID, sender.ID)
		if err != nil {
			return view, merror.Transform(err).Describe("fetching list of actors")
		}

		if err := eventscounts.Incr(ctx, bs.redConn, identities, event.BoxID); err != nil {
			// we log the error but we don’t return it
			logger.FromCtx(ctx).Warn().Err(err).Msg("could not increment new events count")
		}
	}

	return events.ToView(event, sender), nil
}
