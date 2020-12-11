package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/atomic"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
)

// BatchCreateEventRequest ...
type BatchCreateEventRequest struct {
	boxID     string
	BatchType string        `json:"batch_type"`
	Events    []*BatchEvent `json:"events"`
}

// BatchEvent ...
type BatchEvent struct {
	Type       string     `json:"type"`
	Content    types.JSON `json:"content"`
	ReferrerID *string    `json:"referrer_id"`
	Extra      null.JSON  `json:"extra"`
}

// BindAndValidate ...
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

// BatchCreateEvent ...
func (app *BoxApplication) BatchCreateEvent(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*BatchCreateEventRequest)
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	// check the box exists and is not closed
	if err := events.MustBoxBeOpen(ctx, app.DB, req.boxID); err != nil {
		return nil, merror.Transform(err).Describe("checking open")
	}

	// start a transaction to handle all event in one context and potentially rollback all of them
	tr, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, merror.Transform(err).Describe("initing transaction")
	}
	defer atomic.SQLRollback(ctx, tr, &err)

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	createdList := make([]events.Event, len(req.Events))
	metadatas := make(map[string]interface{}, len(req.Events))
	for i, batchE := range req.Events {
		var event events.Event
		event, err = events.New(batchE.Type, batchE.Content, req.boxID, acc.IdentityID, batchE.ReferrerID)
		if err != nil {
			return nil, err
		}

		// call the proper event handlers
		handler := events.Handler(event.Type)

		metadatas[event.ID], err = handler.Do(ctx, &event, batchE.Extra, tr, app.RedConn, identityMapper, app.cryptoRepo, app.filesRepo)
		if err != nil {
			return nil, merror.Transform(err).Describef("doing %s event", event.Type)
		}
		createdList[i] = event
	}

	// handle post-batching action according to the batch type
	if req.BatchType == "accesses" {
		var kicks []events.Event
		kicks, err = events.KickDeprecatedMembers(ctx, tr, identityMapper, req.boxID, acc.IdentityID)
		if err != nil {
			return nil, merror.Transform(err).Describe("potentially kicking")
		}
		createdList = append(createdList, kicks...)
	}

	if cErr := tr.Commit(); cErr != nil {
		return nil, merror.Transform(cErr).Describe("committing transaction")
	}

	// not important to wait for after handlers to return
	// NOTE: we construct a new context since the actual one will be destroyed after the function has returned
	subCtx := context.WithValue(oidc.SetAccesses(context.Background(), acc), logger.CtxKey{}, logger.FromCtx(ctx))
	go func(ctx context.Context, list []events.Event) {
		for _, e := range list {
			for _, after := range events.Handler(e.Type).After {
				if err := after(ctx, &e, app.DB, app.RedConn, identityMapper, app.filesRepo, metadatas[e.ID]); err != nil {
					// we log the error but we don’t return it
					logger.FromCtx(ctx).Warn().Err(err).Msgf("after %s event", e.Type)
				}
			}
		}
	}(subCtx, createdList)

	// build views
	views := make([]events.View, len(createdList))
	var fErr error
	for i, e := range createdList {
		// non-transparent mode
		views[i], fErr = e.Format(ctx, identityMapper, false)
		if err != nil {
			return nil, merror.Transform(fErr).Describe("computing event view")
		}
	}

	return views, nil
}
