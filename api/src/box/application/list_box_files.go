package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/box/events"
)

type ListBoxFilesRequest struct {
	boxID  string
	Offset *int `query:"offset" json:"-"`
	Limit  *int `query:"limit" json:"-"`
}

func (req *ListBoxFilesRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriPath)
	}
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.Offset, v.Min(0)),
		v.Field(&req.Limit, v.Min(0)),
	)
}

func (app *BoxApplication) ListBoxFiles(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListBoxFilesRequest)
	// init an identity mapper for the operation
	identityMapper := app.NewIM()

	acc := oidc.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}
	if err := events.MustMemberHaveAccess(ctx, app.DB, app.RedConn, identityMapper, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	boxEvents, err := events.ListFilesForMembersByBoxID(ctx, app.DB, req.boxID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	fileEvents := make([]*events.Event, len(boxEvents))
	for i := range boxEvents {
		fileEvents[i] = &boxEvents[i]
	}

	if len(fileEvents) != 0 {
		if err := events.SetSavedStatus(ctx, app.DB, acc.IdentityID, fileEvents); err != nil {
			return nil, merror.Transform(err).Describe("setting saved status")
		}
	}

	views := make([]events.View, len(boxEvents))
	for i, e := range boxEvents {
		if err := events.BuildAggregate(ctx, app.DB, &e); err != nil {
			return views, merror.Transform(err).Describe("building aggregate")
		}

		// non-transparent mode to list the event stream
		views[i], err = e.Format(ctx, identityMapper, false)
		if err != nil {
			return views, merror.Transform(err).Describe("computing event view")
		}
	}

	return views, nil
}
