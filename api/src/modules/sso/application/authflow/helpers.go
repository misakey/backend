package authflow

import (
	"net/url"

	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

// BuildRedirectErr with an error code and a description
func buildRedirectErr(code merror.Code, desc string, baseURL *url.URL) string {
	redirectURL := *baseURL
	query := redirectURL.Query()
	// query parameters tends toward compliancy with https://tools.ietf.org/html/rfc6749#section-5.2
	query.Add("error", string(code))
	// TODO: depreciate error_code when frontend is ready
	query.Add("error_code", string(code))
	query.Add("error_description", desc)
	redirectURL.RawQuery = query.Encode()

	return redirectURL.String()
}
