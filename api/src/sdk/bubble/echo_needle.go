package bubble

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// EchoNeedle ...
type EchoNeedle struct {
}

// Explode ...
func (n EchoNeedle) Explode(err error) error {
	echoErr, ok := merr.Cause(err).(*echo.HTTPError)
	if !ok {
		return nil
	}

	// by default we know nothing about the error
	errCode := merr.NoCodeCode
	var desc string

	// act according to error http code:
	// - 400 means echo detected a bad request
	// - 403 means echo detected a permission problem (ex: CSRF)
	// - 404 means echo router did not find the route
	// - 405 means echo router did not find method for requested verb
	switch echoErr.Code {
	case http.StatusBadRequest:
		errCode = merr.BadRequestCode
	case http.StatusUnauthorized:
		errCode = merr.UnauthorizedCode
	case http.StatusNotFound:
		errCode = merr.NotFoundCode
	case http.StatusForbidden:
		errCode = merr.ForbiddenCode
	case http.StatusMethodNotAllowed:
		errCode = merr.MethodNotAllowedCode
	}

	// handle many ways for echo error to express error description
	details := make(map[string]string)
	if echoErr.Internal != nil {
		desc = echoErr.Internal.Error()
	} else if echoErr.Message != nil {
		// we handle some common cases
		switch echoErr.Message {
		case "invalid csrf token":
			details["csrf_token"] = merr.DVInvalid
		}
		desc = fmt.Sprintf("%v", echoErr.Message)
	}

	// final transformation of echo error into merr
	mErr := merr.From(err).Code(errCode).Desc(desc)
	for key, value := range details {
		_ = mErr.Add(key, value)
	}
	return mErr
}
