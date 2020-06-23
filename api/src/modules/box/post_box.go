package box

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/sqlboiler/boil"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

type boxCreationRequest struct {
	PublicKey string `json:"public_key"`
	Title     string `json:"title"`
}

// Validate validates the shape of a box creation request
func (req boxCreationRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.PublicKey, validation.Required),
		validation.Field(&req.Title, validation.Required, validation.Length(5, 50)),
	)
}

func (h *handler) CreateBox(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	// retrieve accesses
	accesses := ajwt.GetAccesses(ctx)
	if accesses == nil {
		return merror.Forbidden()
	}

	// bind and validate the request body
	req := &boxCreationRequest{}
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}

	event, err := events.NewCreate(req.Title, req.PublicKey, accesses.Subject)
	if err != nil {
		return merror.Transform(err).Describe("creating create event")
	}

	// persist the event in storage
	err = event.ToSqlBoiler().Insert(ctx, h.repo.DB(), boil.Infer())
	if err != nil {
		return merror.Transform(err).Describe("inserting event")
	}

	// build the box view and return it
	box, err := events.ComputeBox(ctx, event.BoxID, h.repo, event)
	if err != nil {
		return merror.Transform(err).Describe("building box")
	}

	return eCtx.JSON(http.StatusCreated, box)
}
