package box

import (
	"net/http"
	"strconv"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

type boxQuery struct {
	boxID string
}

func (q boxQuery) Validate() error {
	return v.ValidateStruct(&q,
		v.Field(&q.boxID, v.Required, is.UUIDv4),
	)
}

func (h *handler) getBox(ctx echo.Context) error {
	// TODO access control
	q := boxQuery{boxID: ctx.Param("id")}
	if err := q.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriPath)
	}
	box, err := events.ComputeBox(ctx.Request().Context(), q.boxID, h.db, h.identityRepo)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, box)
}

type boxesQuery struct {
	Offset int `query:"offset" json:"-"`
	Limit  int `query:"limit" json:"-"`
}

func (q boxesQuery) Validate() error {
	return v.ValidateStruct(&q,
		v.Field(&q.Offset, v.Min(0)),
		v.Field(&q.Limit, v.Min(0)),
	)
}

func (h *handler) countBoxes(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	// retrieve accesses to filters boxes to return
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}

	count, err := events.CountSenderBoxes(ctx, h.db, acc.Subject)
	if err != nil {
		return merror.Transform(err).Describe("counting sender boxes")
	}

	eCtx.Response().Header().Set("X-Total-Count", strconv.Itoa(count))
	return eCtx.NoContent(http.StatusNoContent)
}

func (h *handler) listBoxes(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	// retrieve query params
	q := boxesQuery{}

	if err := eCtx.Bind(&q); err != nil {
		return err
	}

	if err := q.Validate(); err != nil {
		return err
	}

	// default limit is 10
	if q.Limit == 0 {
		q.Limit = 10
	}

	// retrieve accesses to filters boxes to return
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}

	boxes, err := events.GetSenderBoxes(
		ctx,
		h.db, h.identityRepo,
		acc.Subject,
		q.Limit, q.Offset,
	)
	if err != nil {
		return merror.Transform(err).Describe("getting sender boxes")
	}

	return eCtx.JSON(http.StatusOK, boxes)
}
