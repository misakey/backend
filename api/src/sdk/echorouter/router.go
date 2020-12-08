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
	// e.Use(middleware.Gzip())                                                          // compress HTTP responses
	// CSRF Token middleware
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-TOKEN",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookiePath:     "/",
	}))
	e.HTTPErrorHandler = errorHandler // custom error handler
	return e
}
