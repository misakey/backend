package generic

import (
	"net/http"

	"github.com/labstack/echo"
)

// genericEcho provides Echo Handler functions interacting with generic operations
type genericEcho struct {
}

// NewGenericEcho is genericEcho constructor
func NewGenericEcho() *genericEcho {
	return &genericEcho{}
}

// Handles version request
func (p *genericEcho) GetVersion(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNoContent)
}
