package merror

import (
	"errors"
)

// Declares some classic errors as variables to use it as a base for our error system
var (
	// ErrBadRequest ...
	ErrBadRequest = errors.New("bad request")
	// ErrUnauthorized ...
	ErrUnauthorized = errors.New("not authorized")
	// ErrForbidden ...
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound ...
	ErrNotFound = errors.New("not found")
	// ErrMethodNotAllowed ...
	ErrMethodNotAllowed = errors.New("method not allowed")
	// ErrConflict ...
	ErrConflict = errors.New("conflict")
	// ErrGone ...
	ErrGone = errors.New("gone")
	// ErrRequestEntityTooLarge ...
	ErrRequestEntityTooLarge = errors.New("request entity too large")
	// ErrUnprocessableEntity ...
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	// ErrClientClosedRequest ...
	ErrClientClosedRequest = errors.New("client closed request")
	// ErrBadGateway ...
	ErrBadGateway = errors.New("bad gateway")
	// ErrServiceUnavailable ...
	ErrServiceUnavailable = errors.New("service unavailable")
	// ErrInternal ...
	ErrInternal = errors.New("internal server")
)

// BadRequest ...
func BadRequest() Error {
	return Error{
		error:   ErrBadRequest,
		Co:      BadRequestCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// Unauthorized ...
func Unauthorized() Error {
	return Error{
		error:   ErrUnauthorized,
		Co:      UnauthorizedCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// Forbidden ...
func Forbidden() Error {
	return Error{
		error:   ErrForbidden,
		Co:      ForbiddenCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// NotFound ...
func NotFound() Error {
	return Error{
		error:   ErrNotFound,
		Co:      NotFoundCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// MethodNotAllowed ...
func MethodNotAllowed() Error {
	return Error{
		error:   ErrMethodNotAllowed,
		Co:      MethodNotAllowedCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// Conflict ...
func Conflict() Error {
	return Error{
		error:   ErrConflict,
		Co:      ConflictCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// RequestEntityTooLarge ...
func RequestEntityTooLarge() Error {
	return Error{
		error:   ErrRequestEntityTooLarge,
		Co:      RequestEntityTooLargeCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// UnprocessableEntity ...
func UnprocessableEntity() Error {
	return Error{
		error:   ErrUnprocessableEntity,
		Co:      UnprocessableEntityCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// ClientClosedRequest ...
func ClientClosedRequest() Error {
	return Error{
		error:   ErrClientClosedRequest,
		Co:      ClientClosedRequestCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// BadGateway ...
func BadGateway() Error {
	return Error{
		error:   ErrBadGateway,
		Co:      BadGatewayCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// ServiceUnavailable ...
func ServiceUnavailable() Error {
	return Error{
		error:   ErrServiceUnavailable,
		Co:      ServiceUnavailableCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// Internal ...
func Internal() Error {
	return Error{
		error:   ErrInternal,
		Co:      InternalCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

// Gone ...
func Gone() Error {
	return Error{
		error:   ErrGone,
		Co:      GoneCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}
