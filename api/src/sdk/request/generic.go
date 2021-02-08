package request

import (
	"context"
	"io"
	"net/http"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"

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

// ResponseStream ...
func ResponseStream(eCtx echo.Context, data interface{}) error {
	readCloser := data.(io.ReadCloser)
	defer func(ctx context.Context) {
		if err := readCloser.Close(); err != nil {
			logger.FromCtx(ctx).Error().Msgf("cannot close stream: %v", err)
		}
	}(eCtx.Request().Context())

	return eCtx.Stream(http.StatusOK, echo.MIMEOctetStream, readCloser)
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
	return eCtx.Blob(http.StatusOK, echo.MIMEOctetStream, data.([]byte))
}

// ResponseRedirectFound ...
func ResponseRedirectFound(eCtx echo.Context, data interface{}) error {
	return eCtx.Redirect(http.StatusFound, data.(string))
}

// HandlerFactory ...
type HandlerFactory struct {
	authzMdlw echo.MiddlewareFunc
}

// NewHandlerFactory ...
func NewHandlerFactory(authzMdlw echo.MiddlewareFunc) HandlerFactory {
	return HandlerFactory{authzMdlw: authzMdlw}
}

// NewPublic ...
func (f *HandlerFactory) NewPublic(
	subPath string,
	initReq func() Request,
	appFunc func(context.Context, Request) (interface{}, error),
	responseFunc func(echo.Context, interface{}) error,
	afterOpts ...func(echo.Context, interface{}) error,
) (string, echo.HandlerFunc) {
	handler := func(eCtx echo.Context) error {
		// directly process the request
		return processReq(eCtx, initReq, appFunc, responseFunc, afterOpts...)
	}
	return subPath, handler
}

// NewOptional ...
func (f *HandlerFactory) NewOptional(
	subPath string,
	initReq func() Request,
	appFunc func(context.Context, Request) (interface{}, error),
	responseFunc func(echo.Context, interface{}) error,
	afterOpts ...func(echo.Context, interface{}) error,
) (string, echo.HandlerFunc, echo.MiddlewareFunc) {
	handler := func(eCtx echo.Context) error {
		// directly process the request
		return processReq(eCtx, initReq, appFunc, responseFunc, afterOpts...)
	}
	return subPath, handler, f.authzMdlw
}

// NewACR3 ...
func (f *HandlerFactory) NewACR3(
	subPath string,
	initReq func() Request,
	appFunc func(context.Context, Request) (interface{}, error),
	responseFunc func(echo.Context, interface{}) error,
	afterOpts ...func(echo.Context, interface{}) error,
) (string, echo.HandlerFunc, echo.MiddlewareFunc) {
	handler := func(eCtx echo.Context) error {
		if err := protectReq(eCtx, oidc.ACR3); err != nil {
			return err
		}
		return processReq(eCtx, initReq, appFunc, responseFunc, afterOpts...)
	}
	return subPath, handler, f.authzMdlw
}

// NewACR2 ...
func (f *HandlerFactory) NewACR2(
	subPath string,
	initReq func() Request,
	appFunc func(context.Context, Request) (interface{}, error),
	responseFunc func(echo.Context, interface{}) error,
	afterOpts ...func(echo.Context, interface{}) error,
) (string, echo.HandlerFunc, echo.MiddlewareFunc) {
	handler := func(eCtx echo.Context) error {
		if err := protectReq(eCtx, oidc.ACR2); err != nil {
			return err
		}
		return processReq(eCtx, initReq, appFunc, responseFunc, afterOpts...)
	}
	return subPath, handler, f.authzMdlw
}

// NewACR1 ...
func (f *HandlerFactory) NewACR1(
	subPath string,
	initReq func() Request,
	appFunc func(context.Context, Request) (interface{}, error),
	responseFunc func(echo.Context, interface{}) error,
	afterOpts ...func(echo.Context, interface{}) error,
) (string, echo.HandlerFunc, echo.MiddlewareFunc) {
	handler := func(eCtx echo.Context) error {
		if err := protectReq(eCtx, oidc.ACR1); err != nil {
			return err
		}
		return processReq(eCtx, initReq, appFunc, responseFunc, afterOpts...)
	}
	return subPath, handler, f.authzMdlw
}

func protectReq(eCtx echo.Context, minACR oidc.ClassRef) error {
	// check accesses if there and acr is compliant
	acc := oidc.GetAccesses(eCtx.Request().Context())
	if acc == nil {
		return merr.Forbidden()
	}
	// check the authentication class reference
	if acc.ACR.LessThan(minACR) {
		return merr.Forbidden().Ori(merr.OriACR).Descf("acr is too low").
			Add("acr", merr.DVForbidden).
			Add("required_acr", minACR.String())
	}
	return nil
}

func processReq(
	eCtx echo.Context,
	initReq func() Request,
	appFunc func(context.Context, Request) (interface{}, error),
	responseFunc func(echo.Context, interface{}) error,
	afterOpts ...func(echo.Context, interface{}) error,
) error {
	ctx := eCtx.Request().Context()

	var req Request
	if initReq != nil {
		req = initReq()
		// bind - validate the request
		if err := req.BindAndValidate(eCtx); err != nil {
			return err
		}
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
