package bubble

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
	// - 403 means echo detected a permission problem (ex: CSRF)
	// - 404 means echo router did not find the route
	// - 405 means echo router did not find method for requested verb
	switch echoErr.Code {
	case http.StatusBadRequest:
		errCode = merror.BadRequestCode
	case http.StatusUnauthorized:
		errCode = merror.UnauthorizedCode
	case http.StatusNotFound:
		errCode = merror.NotFoundCode
	case http.StatusForbidden:
		errCode = merror.ForbiddenCode
	case http.StatusMethodNotAllowed:
		errCode = merror.MethodNotAllowedCode
	}

	// handle many ways for echo error to express error description
	details := make(map[string]string)
	if echoErr.Internal != nil {
		desc = echoErr.Internal.Error()
	} else if echoErr.Message != nil {
		// we handle some common cases
		switch echoErr.Message {
		case "invalid csrf token":
			details["csrf_token"] = merror.DVInvalid
		}
		desc = fmt.Sprintf("%v", echoErr.Message)
	}

	// final transformation of echo error into merror
	mErr := merror.Transform(err).Code(errCode).Describe(desc)
	for key, value := range details {
		_ = mErr.Detail(key, value)
	}
	return mErr
}
