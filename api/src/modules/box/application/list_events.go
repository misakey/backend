package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/request"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

type ListEventsRequest struct {
	boxID  string
	Offset *int `query:"offset" json:"-"`
	Limit  *int `query:"limit" json:"-"`
}

func (req *ListEventsRequest) BindAndValidate(eCtx echo.Context) error {
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

func (bs *BoxApplication) ListEvents(ctx context.Context, genReq request.Request) (interface{}, error) {
	req := genReq.(*ListEventsRequest)

	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return nil, merror.Unauthorized()
	}
	if err := events.MustMemberHaveAccess(ctx, bs.DB, bs.RedConn, bs.Identities, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	boxEvents, err := events.ListForMembersByBoxID(ctx, bs.DB, req.boxID, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	sendersMap, err := events.MapSenderIdentities(ctx, boxEvents, bs.Identities)
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving events senders")
	}
	views := make([]events.View, len(boxEvents))
	for i, e := range boxEvents {
		views[i], err = events.FormatEvent(e, sendersMap)
		if err != nil {
			return views, merror.Transform(err).Describe("computing event view")
		}
	}

	return views, nil
}
