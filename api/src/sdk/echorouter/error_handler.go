package echorouter

import (
	"encoding/json"

	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/bubble"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
)

// Error implements the echo framework error interface.
// it does return error as a JSON object rather than just text.
func Error(err error, ctx echo.Context) {
	// bubble adds more info to the error if known
	err = bubble.Explode(err)
	// force transform to merror and try to interpret the code
	code, mErr := merror.HandleErr(err)

	// log the error
	if mErr.Co == merror.InternalCode {
		details, _ := json.Marshal(mErr.Details)
		logger.FromCtx(ctx.Request().Context()).Error().RawJSON("details", details).Msg(mErr.Desc)
		// flush the internal error information on production to avoid giving too much
		// information to the client
		if env == "production" {
			mErr = mErr.Flush()
		}
	} else {
		details, _ := json.Marshal(mErr.Details)
		logger.FromCtx(ctx.Request().Context()).Info().RawJSON("details", details).Msg(mErr.Desc)
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
