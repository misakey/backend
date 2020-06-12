package box

import (
	"net/http"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

func (h *handler) getBox(ctx echo.Context) error {
	// TODO access control

	boxID := ctx.Param("id")
	if err := v.Validate(boxID, v.Required, is.UUIDv4); err != nil {
		return merror.Transform(err).From(merror.OriPath)
	}
	box, err := events.ComputeBox(ctx.Request().Context(), boxID, h.db, h.identityRepo)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, box)
}
