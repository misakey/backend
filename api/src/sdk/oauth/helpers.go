package oauth

import (
	"net/http"
	"net/url"
	"strings"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// BuildRedirectErr with an error code and a description
func BuildRedirectErr(code merror.Code, desc string, redirectURL *url.URL) string {
	query := redirectURL.Query()
	// query parameters tends toward compliancy with https://tools.ietf.org/html/rfc6749#section-5.2
	query.Add("error", string(code))
	// TODO: depreciate error_code when frontend is ready
	query.Add("error_code", string(code))
	query.Add("error_description", desc)
	redirectURL.RawQuery = query.Encode()

	return redirectURL.String()
}

// redirectErr forging an url with give code and description as query parameters
func (acf *AuthorizationCodeFlow) redirectErr(w http.ResponseWriter, code string, desc string) {
	redirectURL, _ := url.Parse(acf.redirectTokenURL)
	query := redirectURL.Query()
	// query parameters tends toward compliancy with https://tools.ietf.org/html/rfc6749#section-5.2
	query.Add("error", code)
	// TODO: depreciate error_code when frontend is ready
	query.Add("error_code", code)
	query.Add("error_description", desc)
	redirectURL.RawQuery = query.Encode()

	// redirect request
	w.Header().Set("Location", redirectURL.String())
	w.WriteHeader(http.StatusFound)
}

// fromSpacedSep use strings.Split to split a spaced separated list string into a slice
// It handles the fact strings.Split return an slice of size 1 containing empty string if the spacedSep is empty
// We return then, an empty slice instead of this default strings.Split behavior
func fromSpacedSep(spacedSep string) []string {
	if len(spacedSep) == 0 {
		return []string{}
	}
	return strings.Split(spacedSep, " ")
}

// containsString returns true if the candidate string is contained inside container
func containsString(container []string, candidate string) bool {
	for _, element := range container {
		if element == candidate {
			return true
		}
	}
	return false
}
