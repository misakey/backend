package merror

import (
	"net/url"
)

// Code describes an error code as a string used internally and by clearer consumer for better error identifications & reactions.
// It represents global codes and corresponds to http status as described [here](https://en.wikipedia.org/wiki/List_of_HTTP_status_codes).
// We use this link as general specifications for errors and not just for http request errors.
type Code string

const (
	// classic codes
	BadRequestCode            Code = "bad_request"              // 400 general bad request
	UnauthorizedCode          Code = "unauthorized"             // 401 a required valid token is missing/malformed/expired
	ForbiddenCode             Code = "forbidden"                // 403 accesses checks failed
	NotFoundCode              Code = "not_found"                // 404 the resource/route has been not found
	MethodNotAllowedCode      Code = "method_not_allowed"       // 405 the method is not supported at a resource
	ConflictCode              Code = "conflict"                 // 409 the action cannot be perform of the resource
	GoneCode                  Code = "gone"                     // 410 ressource does not exist any more
	RequestEntityTooLargeCode Code = "request_entity_too_large" // 413 the requested entity is too large
	UnprocessableEntityCode   Code = "unprocessable_entity"     // 422 the received entity is unprocessable
	ClientClosedRequestCode   Code = "client_closed_requiest"   // 499 client closed the request so context has been canceled
	InternalCode              Code = "internal"                 // 500 something internal to the service has failed
	BadGatewayCode            Code = "bad_gateway"              // 502 invalid response from the server
	ServiceUnavailableCode    Code = "service_unavailable"      // 503 there service unavailable to handle at this oment

	// no_code codes
	UnknownCode Code = "unknown_code" // 500 something internal to the service has failed
	NoCodeCode  Code = "no_code"      // xxx no specific code defined

	// redirect codes
	AuthProcessRequiredCode Code = "auth_process_required"
	ConsentRequiredCode     Code = "consent_required"
	LoginRequiredCode       Code = "login_required"
	InvalidFlowCode         Code = "invalid_flow"
	MissingParameter        Code = "missing_parameter"
)

func (c Code) String() string {
	return string(c)
}

// AddCodeToURL takes a request and adds a merror.Code to it as a query params
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

// HasCode transforms input error into merror and checks if code are matching
func HasCode(err error, code Code) bool {
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
