package echorouter

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
)

const headerFapiInteractionID = "x-fapi-interaction-id"

// newLogger instantiates and returns a new Logger MiddlewareFunc for echo
func newLogger() echo.MiddlewareFunc {
	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: logger.GetEchoFormat(),
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/version"
		},
	})
}

// newZerologLogger instantiates a zerolog Logger and add it to context
func newZerologLogger(logLevel string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// generate an unique ID for the request
			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				// generate new UUID for the request ID
				id, err := uuid.NewRandom()
				if err != nil {
					c.Error(err)
				}
				requestID = id.String()
			}
			res.Header().Set(echo.HeaderXRequestID, requestID)
			// bind potentially the unique ID to the fapi interaction Header
			// from https://openid.net/specs/openid-financial-api-part-1-ID2.html#protected-resources-provisions
			fapiInterID := req.Header.Get(headerFapiInteractionID)
			if fapiInterID == "" {
				fapiInterID = requestID
			}
			res.Header().Set(headerFapiInteractionID, fapiInterID)
			req.Header.Set(headerFapiInteractionID, fapiInterID)

			// set some values added to all logs
			l := logger.
				ZerologLogger(logLevel).
				With().
				Str("request_id", requestID).
				Str("fapi_interaction_id", fapiInterID).
				Str("protocol", "http").
				Str("uri_path", req.URL.Path).
				Str("method", req.Method).
				Int64("request_bytes", req.ContentLength).
				Logger()

			// replace current request context with a child context containing our zerolog logger
			c.SetRequest(req.WithContext(context.WithValue(req.Context(), logger.CtxKey{}, &l)))
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			return err
		}
	}
}
