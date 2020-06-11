package boxes

import (
	"context"
	"database/sql"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes/repositories/sqlboiler"
)

type PostEventRequest struct {
	events.UserSetFields
}

func (h *handler) postEvent(ctx echo.Context) error {
	accesses := ajwt.GetAccesses(ctx.Request().Context())
	if accesses == nil {
		return merror.Forbidden()
	}

	req := &PostEventRequest{}
	if err := ctx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	boxID := ctx.Param("id")
	err := validation.Validate(boxID, validation.Required, is.UUIDv4)
	if err != nil {
		return merror.Transform(err).Code(merror.BadRequestCode).From(merror.OriPath)
	}

	// "New" performs some shape validation
	event, err := events.New(req.UserSetFields, boxID)
	if err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	boxExists, err := checkBoxExists(ctx.Request().Context(), boxID, h.DB)
	if err != nil {
		return merror.Transform(err).Describe("checking existence of box")
	}
	if !boxExists {
		return merror.NotFound().Describef("no box with id %s", boxID)
	}

	// TODO business logic validation:
	// - access control
	// - for messages, check box is not closed
	// - for box closing, check box is not already closed

	sender, err := h.IdentityService.GetIdentity(ctx.Request().Context(), accesses.Subject)
	if err != nil {
		return merror.Transform(err).Describe("fetching sender identity")
	}
	event.SenderID = sender.ID

	err = event.ToSqlBoiler().Insert(ctx.Request().Context(), h.DB, boil.Infer())
	if err != nil {
		return merror.Transform(err).Describe("inserting event in DB")
	}

	eventView := events.ToView(event, sender)

	return ctx.JSON(http.StatusCreated, eventView)
}

func checkBoxExists(ctx context.Context, boxId string, db *sql.DB) (bool, error) {
	_, err := sqlboiler.Events(
		sqlboiler.EventWhere.BoxID.EQ(boxId),
		sqlboiler.EventWhere.Type.EQ("create"),
	).One(ctx, db)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		} else {
			return false, merror.Transform(err).Describe("retrieving box creation event")
		}
	}
	return true, nil
}
