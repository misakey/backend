package sdk

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

type EchoNeedle struct {
}

func (n EchoNeedle) Explode(err error) error {
	echoErr, ok := merror.Cause(err).(*echo.HTTPError)
	if !ok {
		return nil
	}

	// by default we know nothing about the error
	errCode := merror.NoCodeCode
	var desc string

	// act according to error http code:
	// - 400 means echo detected a bad request
	// - 404 means echo router did not find the route
	// - 405 means echo router did not find method for requested verb
	switch echoErr.Code {
	case http.StatusBadRequest:
		errCode = merror.BadRequestCode
	case http.StatusUnauthorized:
		errCode = merror.UnauthorizedCode
	case http.StatusNotFound:
		errCode = merror.NotFoundCode
	case http.StatusMethodNotAllowed:
		errCode = merror.MethodNotAllowedCode
	}

	// handle many ways for echo error to express error description
	if echoErr.Internal != nil {
		desc = echoErr.Internal.Error()
	} else if echoErr.Message != nil {
		desc = fmt.Sprintf("%v", echoErr.Message)
	}

	// final transformation of echo error into merror
	return merror.Transform(err).Code(errCode).Describe(desc)
}
