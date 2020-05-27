package boxes

import (
	"encoding/json"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes/events"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/boxes/utils"
	"gitlab.misakey.dev/misakey/backend/api/src/sqlboiler"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type creationRequest = boxState

// Validate validates the shape of a box creation request
func (req creationRequest) Validate() error {
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.PublicKey, validation.Required, validation.Match(utils.RxUnpaddedURLsafeBase64)),
		validation.Field(&req.Title, validation.Required, validation.Length(5, 50)),
	); err != nil {
		return err
	}

	return nil
}

func (h *handler) CreateBox(ctx echo.Context) error {
	// TODO: access control

	req := &creationRequest{}
	if err := ctx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	if err := req.Validate(); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	boxID, err := utils.RandomUUIDString()
	if err != nil {
		return merror.Transform(err).Describe("could not generate box ID")
	}

	creationTime := time.Now()

	box := &Box{
		ID:        boxID,
		CreatedAt: creationTime,
		boxState:  *req,
	}

	creationEvent, err := createCreationEvent(req, boxID, creationTime)
	if err != nil {
		return merror.Transform(err).Describe("box creation event")
	}
	err = creationEvent.Insert(ctx.Request().Context(), h.DB, boil.Infer())
	if err != nil {
		return merror.Transform(err)
	}

	return ctx.JSON(http.StatusCreated, box)
}

func createCreationEvent(req *creationRequest, boxID string, creationTime time.Time) (*sqlboiler.Event, error) {
	e := events.Event{}
	e.Type = "create"

	var err error
	e.Content, err = json.Marshal(req)
	if err != nil {
		return nil, merror.Transform(err).From(merror.OriBody).Describe("could not marshall request body")
	}

	e.ID, err = utils.RandomUUIDString()
	if err != nil {
		return nil, merror.Transform(err).Describe("could not generate id for creation event")
	}

	e.BoxID = boxID

	e.CreatedAt = creationTime

	return e.ToSqlBoiler(), nil
}
