package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
)

type CreateEventRequest struct {
	boxID      string
	Type       string     `json:"type"`
	Content    types.JSON `json:"content"`
	ReferrerID *string    `json:"referrer_id"`
}

func (req *CreateEventRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.Type, v.Required, v.In(
			etype.Statelifecycle,
			etype.Msgtext,
			etype.Msgfile,
			etype.Msgedit,
			etype.Msgdelete,
			etype.Memberjoin,
			etype.Memberleave,
		)),
		v.Field(&req.ReferrerID, is.UUIDv4),
		v.Field(&req.Content, v.When(etype.RequiresContent(req.Type), v.Required)),
	)
}

func (bs *BoxApplication) CreateEvent(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*CreateEventRequest)
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	view := events.View{}

	// check the box exists and is not closed
	if err := events.MustBoxBeOpen(ctx, bs.db, req.boxID); err != nil {
		return view, merror.Transform(err).Describe("checking open")
	}

	// init the event
	event, err := events.New(req.Type, req.Content, req.boxID, acc.IdentityID, req.ReferrerID)
	if err != nil {
		return nil, err
	}

	// call the proper event handlers
	handler := events.Handler(event.Type)
	for _, do := range handler.Do {
		if err := do(ctx, &event, bs.db, bs.redConn, bs.identities); err != nil {
			return nil, merror.Transform(err).Describef("during %s event", event.Type)
		}
	}

	// TODO (code structure): use handlers
	if event.Type == "msg.delete" {
		return bs.deleteMessage(ctx, event, handler)
	}
	// TODO (code structure): use handlers
	if event.Type == "msg.edit" {
		return bs.editMessage(ctx, event, handler)
	}

	// not important to wait for after handlers to return
	// NOTE: we construct a new context since the actual one will be destroyed after the function has returned
	subCtx := context.WithValue(ajwt.SetAccesses(context.Background(), acc), logger.CtxKey{}, logger.FromCtx(ctx))
	go func(ctx context.Context, e events.Event) {
		for _, after := range handler.After {
			if err := after(ctx, &e, bs.db, bs.redConn, bs.identities); err != nil {
				// we log the error but we don’t return it
				logger.FromCtx(ctx).Warn().Err(err).Msgf("after %s event", e.Type)
			}
		}
	}(subCtx, event)

	identityMap, err := events.MapSenderIdentities(ctx, []events.Event{event}, bs.identities)
	if err != nil {
		return view, merror.Transform(err).Describe("retrieving identities for view")
	}

	view, err = events.FormatEvent(event, identityMap)
	if err != nil {
		return view, merror.Transform(err).Describe("computing event view")
	}

	return view, nil
}
