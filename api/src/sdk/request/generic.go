package request

import (
	"context"
	"io"
	"net/http"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type Request interface {
	BindAndValidate(echo.Context) error
}

func ResponseNoContent(eCtx echo.Context, _ interface{}) error {
	return eCtx.NoContent(http.StatusNoContent)
}

func ResponseStream(eCtx echo.Context, data interface{}) error {
	readCloser := data.(io.ReadCloser)
	defer func(ctx context.Context) {
		if err := readCloser.Close(); err != nil {
			logger.FromCtx(ctx).Error().Msgf("cannot close stream: %v", err)
		}
	}(eCtx.Request().Context())

	return eCtx.Stream(http.StatusOK, echo.MIMEOctetStream, readCloser)
}

func ResponseOK(eCtx echo.Context, data interface{}) error {
	return eCtx.JSON(http.StatusOK, data)
}

func ResponseCreated(eCtx echo.Context, data interface{}) error {
	return eCtx.JSON(http.StatusCreated, data)
}

func ResponseBlob(eCtx echo.Context, data interface{}) error {
	return eCtx.Blob(http.StatusOK, echo.MIMEOctetStream, data.([]byte))
}

func NewPublicHTTP(
	initReq func() Request,
	appFunc func(context.Context, Request) (interface{}, error),
	responseFunc func(echo.Context, interface{}) error,
	afterOpts ...func(echo.Context, interface{}) error,
) func(eCtx echo.Context) error {
	return newHTTPEntrypoint(initReq, false, appFunc, responseFunc, afterOpts...)
}

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
			if acc := ajwt.GetAccesses(ctx); acc == nil {
				return merror.Forbidden()
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
