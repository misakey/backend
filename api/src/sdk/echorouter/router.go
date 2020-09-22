package echorouter

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var env = os.Getenv("ENV")

func New() *echo.Echo {
	e := echo.New()
	e.Use(NewZerologLogger())
	e.Use(NewLogger())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = Error
	e.HideBanner = true
	return e
}
