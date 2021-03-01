package echorouter

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var env = os.Getenv("ENV")

// New ...
func New(logLevel string) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(newZerologLogger(logLevel))                                                 // init contextual logger for the request
	e.Use(newLogger())                                                                // log received requests
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{StackSize: 4 << 16})) // increase default stack size to always print it fully
	// e.Use(middleware.Gzip()) // compress HTTP responses
	e.HTTPErrorHandler = errorHandler // custom error handler

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-TOKEN",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookiePath:     "/",
		Skipper: func(ctx echo.Context) bool {
			// csrf is ignored on request using headers - machine only should do that, this is checked after token introspection
			return ctx.Request().Header.Get("Authorization") != ""
		},
	}))
	return e
}
