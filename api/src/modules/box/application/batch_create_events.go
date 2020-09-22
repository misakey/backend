package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events/etype"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
)

type BatchCreateEventRequest struct {
	boxID     string
	BatchType string        `json:"batch_type"`
	Events    []*BatchEvent `json:"events"`
}

type BatchEvent struct {
	Type       string     `json:"type"`
	Content    types.JSON `json:"content"`
	ReferrerID *string    `json:"referrer_id"`
}

func (req *BatchCreateEventRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("id")
	if err := v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.BatchType, v.Required, v.In("accesses")),
		v.Field(&req.Events, v.Required),
	); err != nil {
		return err
	}
	return v.Validate(req.Events)
}

// Validate is a separated declared function so BatchEvent implements v.Validatable interface
// then BatchCreateEvent can use v.Each to validate them
func (req BatchEvent) Validate() error {
	return v.ValidateStruct(&req,
		v.Field(&req.Type, v.Required, v.In(etype.Accessadd, etype.Accessrm)),
		v.Field(&req.ReferrerID, is.UUIDv4),
		v.Field(&req.Content, v.When(etype.RequiresContent(req.Type), v.Required)),
	)
}

func (bs *BoxApplication) BatchCreateEvent(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*BatchCreateEventRequest)
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	// check the box exists and is not closed
	if err := events.MustBoxBeOpen(ctx, bs.db, req.boxID); err != nil {
		return nil, merror.Transform(err).Describe("checking open")
	}

	// start a transaction to handle all event in one context and potentially rollback all of them
	tr, err := bs.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, merror.Transform(err).Describe("initing transaction")
	}
	defer atomic.SQLRollback(ctx, tr, err)

	createdList := make([]events.Event, len(req.Events))
	for i, batchE := range req.Events {
		var event events.Event
		event, err = events.New(batchE.Type, batchE.Content, req.boxID, acc.IdentityID, batchE.ReferrerID)
		if err != nil {
			return nil, err
		}

		// call the proper event handlers
		handler := events.Handler(event.Type)

		for _, do := range handler.Do {
			err = do(ctx, &event, tr, bs.redConn, bs.identities)
			if err != nil {
				return nil, merror.Transform(err).Describef("doing %s event", event.Type)
			}
		}
		createdList[i] = event
	}

	// handle post-batching action according to the batch type
	if req.BatchType == "accesses" {
		var kicks []events.Event
		kicks, err = events.KickDeprecatedMembers(ctx, tr, bs.identities, req.boxID, acc.IdentityID)
		if err != nil {
			return nil, merror.Transform(err).Describe("potentially kicking")
		}
		createdList = append(createdList, kicks...)
	}

	err = tr.Commit()
	if err != nil {
		return nil, merror.Transform(err).Describe("committing transaction")
	}

	for _, event := range createdList {
		for _, after := range events.Handler(event.Type).After {
			if err := after(ctx, &event, bs.db, bs.redConn, bs.identities); err != nil {
				// we log the error but we don’t return it
				logger.FromCtx(ctx).Warn().Err(err).Msgf("after %s event", event.Type)
			}
		}
	}

	sendersMap, err := events.MapSenderIdentities(ctx, createdList, bs.identities)
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving events senders")
	}
	views := make([]events.View, len(createdList))
	for i, e := range createdList {
		views[i], err = events.FormatEvent(e, sendersMap)
		if err != nil {
			return nil, merror.Transform(err).Describe("computing event view")
		}
	}

	return views, nil
}
