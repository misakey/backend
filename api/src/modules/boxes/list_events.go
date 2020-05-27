package boxes

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes/events"
	"gitlab.misakey.dev/misakey/backend/api/src/sqlboiler"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

func (h *handler) listEvents(ctx echo.Context) error {
	// TODO access control

	boxID := ctx.Param("id")
	err := validation.Validate(boxID, validation.Required, is.UUIDv4)
	if err != nil {
		return merror.Transform(err).Code(merror.BadRequestCode).From(merror.OriPath)
	}

	dbEvents, err := sqlboiler.Events(
		sqlboiler.EventWhere.BoxID.EQ(boxID),
		qm.OrderBy(sqlboiler.EventColumns.CreatedAt),
	).All(ctx.Request().Context(), h.DB)
	if err != nil {
		return merror.Transform(err).Describe("retrieving events")
	}

	if dbEvents == nil {
		return merror.NotFound().Describef("no box with id %s", boxID)
	}

	var result []*events.Event
	for _, e := range dbEvents {
		result = append(result, events.FromSqlBoiler(e))
	}

	return ctx.JSON(http.StatusOK, result)
}
