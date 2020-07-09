package box

import (
	"net/http"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

type ackNewEventsCountRequest struct {
	boxID string

	IdentityID string `json:"identity_id"`
}

func (req ackNewEventsCountRequest) Validate() error {
	return v.ValidateStruct(&req,
		v.Field(&req.boxID, v.Required, is.UUIDv4),
		v.Field(&req.IdentityID, v.Required, is.UUIDv4),
	)
}

func (h *handler) ackNewEventsCount(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()

	// retrieve accesses
	acc := ajwt.GetAccesses(ctx)
	if acc == nil {
		return merror.Forbidden()
	}

	// bind - validate the request
	req := &ackNewEventsCountRequest{}
	if err := eCtx.Bind(req); err != nil {
		return merror.Transform(err).From(merror.OriBody)
	}
	req.boxID = eCtx.Param("id")

	if err := req.Validate(); err != nil {
		return err
	}

	if acc.Subject != req.IdentityID {
		return merror.Forbidden()
	}

	if err := h.repo.EventsCounts().Del(ctx, req.IdentityID, req.boxID); err != nil {
		return err
	}

	return eCtx.NoContent(http.StatusNoContent)
}
