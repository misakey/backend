package merror

import (
	"log"
	"net/http"
)

var (
	// StatusClientClosedRequest shall never be returned to consumer because of the context of the code
	StatusClientClosedRequest = 499 // http://lxr.nginx.org/source/src/http/ngx_http_request.h#0120
)

// ToHTTPCode returns HTTP code corresponding to Domain Misakey Code
func ToHTTPCode(err error) int {
	mErr := transform(err)
	switch mErr.Co {
	case UnauthorizedCode:
		return http.StatusUnauthorized
	case BadRequestCode:
		return http.StatusBadRequest
	case ForbiddenCode:
		return http.StatusForbidden
	case NotFoundCode:
		return http.StatusNotFound
	case MethodNotAllowedCode:
		return http.StatusMethodNotAllowed
	case ConflictCode:
		return http.StatusConflict
	case GoneCode:
		return http.StatusGone
	case RequestEntityTooLargeCode:
		return http.StatusRequestEntityTooLarge
	case UnprocessableEntityCode:
		return http.StatusUnprocessableEntity
	case ClientClosedRequestCode:
		return StatusClientClosedRequest
	case BadGatewayCode:
		return http.StatusBadGateway
	case ServiceUnavailableCode:
		return http.StatusServiceUnavailable
	}
	return http.StatusInternalServerError
}

// TransformHTTPCode returns merror corresponding to HTTP code
func TransformHTTPCode(code int) Error {
	switch code {
	case http.StatusBadRequest:
		return BadRequest()
	case http.StatusUnauthorized:
		return Unauthorized()
	case http.StatusForbidden:
		return Forbidden()
	case http.StatusNotFound:
		return NotFound()
	case http.StatusMethodNotAllowed:
		return MethodNotAllowed()
	case http.StatusConflict:
		return Conflict()
	case http.StatusRequestEntityTooLarge:
		return RequestEntityTooLarge()
	case http.StatusUnprocessableEntity:
		return UnprocessableEntity()
	case StatusClientClosedRequest:
		return ClientClosedRequest()
	case http.StatusBadGateway:
		return BadGateway()
	case http.StatusServiceUnavailable:
		return ServiceUnavailable()
	}
	return Internal()
}

// HandleErr tries to interpret an error as a Misakey Error returning HTTTP Code aside it.
// Its set default value if the error is not a Misakey Error
func HandleErr(err error) (int, Error) {
	var finalErr Error

	// try to use what has been sent
	mErr, ok := err.(Error)
	if ok {
		finalErr = mErr
	} else {
		finalErr = Internal().Describe(err.Error())
	}

	// check Error clearness - panic if not clear
	if !finalErr.Clear() {
		log.Fatalf("an error is not clear (%s - %s - %s)!", finalErr.Co, finalErr.Desc, finalErr.Error())
	}
	return ToHTTPCode(finalErr), finalErr
}
