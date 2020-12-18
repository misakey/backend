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

// CreateEventRequest ...
type CreateEventRequest struct {
	boxID               string
	Type                string     `json:"type"`
	Content             types.JSON `json:"content"`
	ReferrerID          *string    `json:"referrer_id"`
	Extra               null.JSON  `json:"extra"`
	MetadataForHandlers events.MetadataForUsedSpaceHandler
}

// BindAndValidate ...
func (req *CreateEventRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.Type, v.Required, v.In(
			etype.StateKeyShare,
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

// CreateEvent ...
func (app *BoxApplication) CreateEvent(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*CreateEventRequest)
	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	view := events.View{}

	// check the box exists
	if err := events.MustBoxExists(ctx, app.DB, req.boxID); err != nil {
		return view, merror.Transform(err).Describe("checking exist")
	}

	// init the event
	event, err := events.New(req.Type, req.Content, req.boxID, acc.IdentityID, req.ReferrerID)
	if err != nil {
		return nil, err
	}
	// used for computing newBoxUsedSpace in handlers
	event.MetadataForHandlers = req.MetadataForHandlers

	// call the proper event handlers
	// initialize transaction
	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, merror.Transform(err).Describe("creating DB transaction")
	}

	handler := events.Handler(event.Type)
	metadata, err := handler.Do(ctx, &event, req.Extra, tx, app.RedConn, identityMapper, app.cryptoRepo, app.filesRepo)
	if err != nil {
		atomic.SQLRollback(ctx, tx, &err)
		return nil, merror.Transform(err).Describef("during %s event", event.Type)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return nil, merror.Transform(err).Describe("committing transaction")
	}

	// not important to wait for after handlers to return
	// NOTE: we construct a new context since the actual one will be destroyed after the function has returned
	subCtx := context.WithValue(oidc.SetAccesses(context.Background(), acc), logger.CtxKey{}, logger.FromCtx(ctx))
	go func(ctx context.Context, e events.Event) {
		for _, after := range handler.After {
			if err := after(ctx, &e, app.DB, app.RedConn, identityMapper, app.filesRepo, metadata); err != nil {
				// we log the error but we don’t return it
				logger.FromCtx(ctx).Warn().Err(err).Msgf("after %s event", e.Type)
			}
		}
	}(subCtx, event)

	// finally format the event (non-transparent mode)
	view, err = event.Format(ctx, identityMapper, false)
	if err != nil {
		return view, merror.Transform(err).Describe("computing event view")
	}

	return view, nil
}
