package echorouter

import (
	"encoding/json"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/bubble"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
)

// errorHandler implements the echo framework error interface.
// it does return error as a JSON object rather than just text.
func errorHandler(err error, ctx echo.Context) {
	// bubble adds more info to the error if known
	err = bubble.Explode(err)
	// force transform to merror and try to interpret the code
	code, mErr := merror.HandleErr(err)

	// log the error
	details, _ := json.Marshal(mErr.Details)
	logEvent := logger.FromCtx(ctx.Request().Context())
	// NOTE: log an error on internal code whereas log an info on others
	if mErr.Co == merror.InternalCode {
		logEvent.Error().RawJSON("details", details).Msg(mErr.Desc)
		// flush the internal error information on production to avoid giving too much
		// information to the client
		if env == "production" {
			mErr = mErr.Flush()
		}
	} else {
		logEvent.Info().RawJSON("details", details).Msg(mErr.Desc)
	}

	if !ctx.Response().Committed {
		// we don't return any body response in case of HEAD request
		if ctx.Request().Method == echo.HEAD {
			_ = ctx.NoContent(code)
		} else {
			_ = ctx.JSON(code, mErr)
		}
	}
}
