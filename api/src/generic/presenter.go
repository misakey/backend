package generic

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// genericEcho provides Echo Handler functions interacting with generic operations
type genericEcho struct {
}

// NewGenericEcho is genericEcho constructor
func newGenericEcho() *genericEcho {
	return &genericEcho{}
}

// Handles version request
func (p *genericEcho) GetVersion(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNoContent)
}

func (p *genericEcho) GetCSRF(ctx echo.Context) error {
	type csrfView struct {
		CSRFToken string `json:"csrf_token"`
	}
	csrfToken := csrfView{
		CSRFToken: fmt.Sprintf("%s", ctx.Get("csrf")),
	}
	return ctx.JSON(http.StatusOK, csrfToken)
}
