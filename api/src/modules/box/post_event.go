package box

import (
	"context"
	"database/sql"
	"net/http"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

type postEventRequest struct {
	boxID string

	Type    string     `json:"type"`
	Content types.JSON `json:"content"`
}

func (req postEventRequest) Validate() error {
	return v.ValidateStruct(&req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.Type, v.Required, v.In("msg.text", "state.lifecycle")),
		v.Field(&req.Content, v.Required),
	)
}

func (h *handler) postEvent(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	// retrieve accesses
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}

	// bind - validate the request
	req := &postEventRequest{}
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("id")

	if err := req.Validate(); err != nil {
		return err
	}

	// "New" performs some shape validation
	event, err := events.New(req.Type, req.Content, req.boxID, acc.Subject)
	if err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	view, err := h.createEvent(ctx, event)
	if err != nil {
		return err
	}
	return eCtx.JSON(http.StatusCreated, view)
}

func (h *handler) createEvent(
	ctx context.Context,
	e events.Event,
) (events.View, error) {
	view := events.View{}

	// check the box does exist
	boxExists, err := checkBoxExists(ctx, e.BoxID, h.repo.DB())
	if err != nil {
		return view, merror.Transform(err).Describe("checking existence of box")
	}
	if !boxExists {
		return view, merror.NotFound().Detail("id", merror.DVNotFound).
			Describef("no box with id %s", e.BoxID)
	}

	// check the box is not closed
	if err := boxes.MustBeOpen(ctx, h.repo.DB(), e.BoxID); err != nil {
		return view, merror.Transform(err).Describe("checking open")
	}

	// check the sender is creator of the box for state.lifecycle events
	if e.Type == "state.lifecycle" {
		if err := boxes.MustBeCreator(ctx, h.repo.DB(), e.BoxID, e.SenderID); err != nil {
			return view, merror.Transform(err).Describe("checking creator")
		}
	}

	sender, err := h.repo.Identities().Get(ctx, e.SenderID)
	if err != nil {
		return view, merror.Transform(err).Describe("fetching sender identity")
	}

	if err := e.ToSQLBoiler().Insert(ctx, h.repo.DB(), boil.Infer()); err != nil {
		return view, merror.Transform(err).Describe("inserting event in DB")
	}

	// increment count for all identities except the sender
	// only if this is not a lifecycle event
	if e.Type != "state.lifecycle" {
		identities, err := boxes.GetActorsExcept(ctx, h.repo.DB(), e.BoxID, sender.ID)
		if err != nil {
			return view, merror.Transform(err).Describe("fetching list of actors")
		}

		if err := h.repo.EventsCounts().Incr(ctx, identities, e.BoxID); err != nil {
			// we log the error but we don’t return it
			logger.FromCtx(ctx).Warn().Err(err).Msg("could not increment new events count")
		}
	}

	return events.ToView(e, sender), nil
}

func checkBoxExists(ctx context.Context, boxID string, exec boil.ContextExecutor) (bool, error) {
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
