package merr

import (
	"net/url"
)

// Code describes an error code as a string used internally and by clearer consumer for better error identifications & reactions.
// It represents global codes and corresponds to http status as described [here](https://en.wikipedia.org/wiki/List_of_HTTP_status_codes).
// We use this link as general specifications for errors and not just for http request errors.
type Code string

// code constants
const (
	// classic codes
	BadRequestCode            Code = "bad_request"
	UnauthorizedCode          Code = "unauthorized"
	ForbiddenCode             Code = "forbidden"
	NotFoundCode              Code = "not_found"
	MethodNotAllowedCode      Code = "method_not_allowed"
	ConflictCode              Code = "conflict"
	GoneCode                  Code = "gone"
	RequestEntityTooLargeCode Code = "request_entity_too_large"
	UnprocessableEntityCode   Code = "unprocessable_entity"
	ClientClosedRequestCode   Code = "client_closed_requiest"
	InternalCode              Code = "internal"
	BadGatewayCode            Code = "bad_gateway"
	ServiceUnavailableCode    Code = "service_unavailable"

	// no_code codes
	UnknownCode Code = "unknown_code"
	NoCodeCode  Code = "no_code"

	// redirect codes
	AuthProcessRequiredCode Code = "auth_process_required"
	ConsentRequiredCode     Code = "consent_required"
	LoginRequiredCode       Code = "login_required"
	InvalidFlowCode         Code = "invalid_flow"
	MissingParameter        Code = "missing_parameter"
)

// String ...
func (c Code) String() string {
	return string(c)
}

// AddCodeToURL takes a request and adds a merr.Code to it as a query params
func AddCodeToURL(req string, code Code) (string, error) {
	requestURL, err := url.ParseRequestURI(req)
	if err != nil {
		return "", err
	}

	// prepare Query Parameters
	params := url.Values{}
	params.Add("error", string(code))
	// TODO: depreciate error_code when frontend is ready
	params.Add("error_code", string(code))

	// add query parameters to the URL
	requestURL.RawQuery = params.Encode() // escape Query Parameters
	return requestURL.String(), nil
}

// hasCode transforms input error into merr and checks if code are matching
// hasCode return false on nil received error
func hasCode(err error, code Code) bool {
	return err != nil && transform(err).Co == code
}

// ToCode takes a default error and return corresponding Code
func ToCode(err error) Code {
	errType := Cause(err)
	switch errType {
	case ErrBadRequest:
		return BadRequestCode
	case ErrUnauthorized:
		return UnauthorizedCode
	case ErrForbidden:
		return ForbiddenCode
	case ErrNotFound:
		return NotFoundCode
	case ErrMethodNotAllowed:
		return MethodNotAllowedCode
	case ErrConflict:
		return ConflictCode
	case ErrRequestEntityTooLarge:
		return RequestEntityTooLargeCode
	case ErrUnprocessableEntity:
		return UnprocessableEntityCode
	case ErrClientClosedRequest:
		return ClientClosedRequestCode
	case ErrServiceUnavailable:
		return ServiceUnavailableCode
	}
	return InternalCode
}

func IsANotFound(err error) bool {
	return hasCode(err, NotFoundCode)
}

func IsAForbidden(err error) bool {
	return hasCode(err, ForbiddenCode)
}

func IsAConflict(err error) bool {
	return hasCode(err, ConflictCode)
}

func IsAnInternal(err error) bool {
	return hasCode(err, InternalCode)
}
