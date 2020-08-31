package application

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/boxes"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/entrypoints"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

type ListEventsRequest struct {
	boxID string
}

func (req *ListEventsRequest) BindAndValidate(eCtx echo.Context) error {
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriPath)
	}
	req.boxID = eCtx.Param("id")
	return v.ValidateStruct(req,
		v.Field(&req.boxID, is.UUIDv4),
	)
}

func (bs *BoxApplication) ListEvents(ctx context.Context, genReq entrypoints.Request) (interface{}, error) {
	req := genReq.(*ListEventsRequest)
	acc := ajwt.GetAccesses(ctx)

	// if the box is closed, only the creator can list its events
	if err := boxes.MustBeCreatorIfClosed(ctx, bs.db, req.boxID, acc.IdentityID); err != nil {
		return nil, err
	}

	// list
	boxEvents, err := events.ListByBoxID(ctx, bs.db, req.boxID)
	if err != nil {
		return nil, err
	}

	sendersMap, err := events.MapSenderIdentities(ctx, boxEvents, bs.identities)
	if err != nil {
		return nil, merror.Transform(err).Describe("retrieving events senders")
	}

	views := make([]events.View, len(boxEvents))
	for i, e := range boxEvents {
		views[i], err = events.ToView(e, sendersMap)
		if err != nil {
			return views, merror.Transform(err).Describe("computing event view")
		}
	}

	return views, nil
}
