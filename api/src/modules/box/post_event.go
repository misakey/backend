package box

import (
	"context"
	"database/sql"
	"net/http"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/types"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

type postEventRequest struct {
	boxID string

	Type    string     `json:"type"`
	Content types.JSON `json:"content"`
}

func (req postEventRequest) Validate() error {
	return v.ValidateStruct(&req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.Type, v.Required, v.In("create", "msg.text", "msg.file", "state.lifecycle")),
		v.Field(&req.Content, v.Required),
	)
}

func (h *handler) postEvent(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	// retrieve accesses
	accesses := ajwt.GetAccesses(ctx)
	if accesses == nil {
		return merror.Forbidden()
	}

	// bind - validate the request
	req := &postEventRequest{}
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("id")

	if err := req.Validate(); err != nil {
		return err
	}

	// "New" performs some shape validation
	event, err := events.New(req.Type, req.Content, req.boxID, accesses.Subject)
	if err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	boxExists, err := checkBoxExists(ctx, req.boxID, h.db)
	if err != nil {
		return merror.Transform(err).Describe("checking existence of box")
	}
	if !boxExists {
		return merror.NotFound().Detail("id", merror.DVNotFound).
			Describef("no box with id %s", req.boxID)
	}

	// TODO business logic validation:
	// - access control
	// - for messages, check box is not closed
	// - for box closing, check box is not already closed

	sender, err := h.identityRepo.GetIdentity(ctx, accesses.Subject)
	if err != nil {
		return merror.Transform(err).Describe("fetching sender identity")
	}

	err = event.ToSqlBoiler().Insert(ctx, h.db, boil.Infer())
	if err != nil {
		return merror.Transform(err).Describe("inserting event in DB")
	}

	eventView := events.ToView(event, sender)
	return eCtx.JSON(http.StatusCreated, eventView)
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
