package entrypoints

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/oidc"
)

// Request ...
type Request interface {
	BindAndValidate(echo.Context) error
}

// ResponseNoContent ...
func ResponseNoContent(eCtx echo.Context, _ interface{}) error {
	return eCtx.NoContent(http.StatusNoContent)
}

// ResponseOK ...
func ResponseOK(eCtx echo.Context, data interface{}) error {
	return eCtx.JSON(http.StatusOK, data)
}

// ResponseCreated ...
func ResponseCreated(eCtx echo.Context, data interface{}) error {
	return eCtx.JSON(http.StatusCreated, data)
}

// ResponseBlob ...
func ResponseBlob(eCtx echo.Context, data interface{}) error {
	return eCtx.Blob(http.StatusOK, "application/octet-stream", data.([]byte))
}

// NewPublicHTTP ...
func NewPublicHTTP(
	initReq func() Request,
	appFunc func(context.Context, Request) (interface{}, error),
	responseFunc func(echo.Context, interface{}) error,
	afterOpts ...func(echo.Context, interface{}) error,
) func(eCtx echo.Context) error {
	return newHTTPEntrypoint(initReq, false, appFunc, responseFunc, afterOpts...)
}

// NewProtectedHTTP ...
func NewProtectedHTTP(
	initReq func() Request,
	appFunc func(context.Context, Request) (interface{}, error),
	responseFunc func(echo.Context, interface{}) error,
	afterOpts ...func(echo.Context, interface{}) error,
) func(eCtx echo.Context) error {
	return newHTTPEntrypoint(initReq, true, appFunc, responseFunc, afterOpts...)
}

func newHTTPEntrypoint(
	initReq func() Request,
	checkAccesses bool,
	appFunc func(context.Context, Request) (interface{}, error),
	responseFunc func(echo.Context, interface{}) error,
	afterOpts ...func(echo.Context, interface{}) error,
) func(eCtx echo.Context) error {
	return func(
		eCtx echo.Context,
	) error {
		req := initReq()
		ctx := eCtx.Request().Context()

		if checkAccesses {
			if acc := oidc.GetAccesses(ctx); acc == nil {
				return merr.Forbidden()
			}
		}

		// bind - validate the request
		if err := req.BindAndValidate(eCtx); err != nil {
			return err
		}

		data, err := appFunc(ctx, req)
		if err != nil {
			return err
		}

		for _, opt := range afterOpts {
			if err := opt(eCtx, data); err != nil {
				return err
			}
		}

		return responseFunc(eCtx, data)
	}
}
