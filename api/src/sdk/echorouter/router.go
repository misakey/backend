package echorouter

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var env = os.Getenv("ENV")

func New(logLevel string) *echo.Echo {
	e := echo.New()
	e.Use(NewZerologLogger(logLevel))
	e.Use(NewLogger())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = Error
	e.HideBanner = true
	return e
}
