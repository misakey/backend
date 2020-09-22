package merror

import (
	"errors"
)

// Declares some classic errors as variables to use it as a base for our error system
var (
	BadRequestError            = errors.New("bad request")
	UnauthorizedError          = errors.New("not authorized")
	ForbiddenError             = errors.New("forbidden")
	NotFoundError              = errors.New("not found")
	MethodNotAllowedError      = errors.New("method not allowed")
	ConflictError              = errors.New("conflict")
	GoneError                  = errors.New("gone")
	RequestEntityTooLargeError = errors.New("request entity too large")
	UnprocessableEntityError   = errors.New("unprocessable entity")
	ClientClosedRequestError   = errors.New("client closed request")
	BadGatewayError            = errors.New("bad gateway")
	ServiceUnavailableError    = errors.New("service unavailable")
	InternalError              = errors.New("internal server")
)

func BadRequest() Error {
	return Error{
		error:   BadRequestError,
		Co:      BadRequestCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func Unauthorized() Error {
	return Error{
		error:   UnauthorizedError,
		Co:      UnauthorizedCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func Forbidden() Error {
	return Error{
		error:   ForbiddenError,
		Co:      ForbiddenCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func NotFound() Error {
	return Error{
		error:   NotFoundError,
		Co:      NotFoundCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func MethodNotAllowed() Error {
	return Error{
		error:   MethodNotAllowedError,
		Co:      MethodNotAllowedCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func Conflict() Error {
	return Error{
		error:   ConflictError,
		Co:      ConflictCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func RequestEntityTooLarge() Error {
	return Error{
		error:   RequestEntityTooLargeError,
		Co:      RequestEntityTooLargeCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func UnprocessableEntity() Error {
	return Error{
		error:   UnprocessableEntityError,
		Co:      UnprocessableEntityCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func ClientClosedRequest() Error {
	return Error{
		error:   ClientClosedRequestError,
		Co:      ClientClosedRequestCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func BadGateway() Error {
	return Error{
		error:   BadGatewayError,
		Co:      BadGatewayCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func ServiceUnavailable() Error {
	return Error{
		error:   ServiceUnavailableError,
		Co:      ServiceUnavailableCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func Internal() Error {
	return Error{
		error:   InternalError,
		Co:      InternalCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}

func Gone() Error {
	return Error{
		error:   GoneError,
		Co:      GoneCode,
		Ori:     OriNotDefined,
		Details: make(map[string]string),
	}
}
