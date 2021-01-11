package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/box/events/etype"
)

// ListEventsRequest ...
type ListEventsRequest struct {
	boxID  string
	Offset *int `query:"offset" json:"-"`
	Limit  *int `query:"limit" json:"-"`
}

// BindAndValidate ...
func (req *ListEventsRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merr.From(err).Ori(merr.OriPath)
	}
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.Offset, v.Min(0)),
		v.Field(&req.Limit, v.Min(0)),
	)
}

// ListEvents ...
func (app *BoxApplication) ListEvents(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListEventsRequest)

	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merr.Unauthorized()
	}
	if err := events.MustBeMember(ctx, app.DB, app.RedConn, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	boxEvents, err := events.ListForMembersByBoxID(ctx, app.DB, req.boxID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	// add information on files
	var fileEvents []*events.Event
	for i, e := range boxEvents {
		if e.Type == etype.Msgfile {
			fileEvents = append(fileEvents, &boxEvents[i])
		}
	}

	if len(fileEvents) != 0 {
		if err := events.SetSavedStatus(ctx, app.DB, acc.IdentityID, fileEvents); err != nil {
			return nil, merr.From(err).Desc("setting saved status")
		}
	}

	// build the returned views
	views := make([]events.View, len(boxEvents))
	for i, e := range boxEvents {

		if err := events.BuildAggregate(ctx, app.DB, &e); err != nil {
			return views, merr.From(err).Desc("building aggregate")
		}

		// non-transparent mode to list the event stream
		views[i], err = e.Format(ctx, identityMapper, false)
		if err != nil {
			return views, merr.From(err).Desc("computing event view")
		}
	}

	return views, nil
}
