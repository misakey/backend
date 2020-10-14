package merror

import (
	"errors"
)

// Declares some classic errors as variables to use it as a base for our error system
var (
	ErrBadRequest            = errors.New("bad request")
	ErrUnauthorized          = errors.New("not authorized")
	ErrForbidden             = errors.New("forbidden")
	ErrNotFound              = errors.New("not found")
	ErrMethodNotAllowed      = errors.New("method not allowed")
	ErrConflict              = errors.New("conflict")
	ErrGone                  = errors.New("gone")
	ErrRequestEntityTooLarge = errors.New("request entity too large")
	ErrUnprocessableEntity   = errors.New("unprocessable entity")
	ErrClientClosedRequest   = errors.New("client closed request")
	ErrBadGateway            = errors.New("bad gateway")
	ErrServiceUnavailable    = errors.New("service unavailable")
	ErrInternal              = errors.New("internal server")
)

func BadRequest() Error {
	return Error{
		error:   ErrBadRequest,
		Co:      BadRequestCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func Unauthorized() Error {
	return Error{
		error:   ErrUnauthorized,
		Co:      UnauthorizedCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func Forbidden() Error {
	return Error{
		error:   ErrForbidden,
		Co:      ForbiddenCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func NotFound() Error {
	return Error{
		error:   ErrNotFound,
		Co:      NotFoundCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func MethodNotAllowed() Error {
	return Error{
		error:   ErrMethodNotAllowed,
		Co:      MethodNotAllowedCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func Conflict() Error {
	return Error{
		error:   ErrConflict,
		Co:      ConflictCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func RequestEntityTooLarge() Error {
	return Error{
		error:   ErrRequestEntityTooLarge,
		Co:      RequestEntityTooLargeCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func UnprocessableEntity() Error {
	return Error{
		error:   ErrUnprocessableEntity,
		Co:      UnprocessableEntityCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func ClientClosedRequest() Error {
	return Error{
		error:   ErrClientClosedRequest,
		Co:      ClientClosedRequestCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func BadGateway() Error {
	return Error{
		error:   ErrBadGateway,
		Co:      BadGatewayCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func ServiceUnavailable() Error {
	return Error{
		error:   ErrServiceUnavailable,
		Co:      ServiceUnavailableCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func Internal() Error {
	return Error{
		error:   ErrInternal,
		Co:      InternalCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func Gone() Error {
	return Error{
		error:   ErrGone,
		Co:      GoneCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}
