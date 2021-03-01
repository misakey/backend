package format

import (
	"net/url"
	"regexp"
)

// UnpaddedURLSafeBase64 ...
var UnpaddedURLSafeBase64 = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
var BearerToken = regexp.MustCompile(`^Bearer ([\S]+)`)

// AddQueryParam to the originalURL using the net/url
// paramKey as query parameter key and paramVal as query parameter value
func AddQueryParam(originalURL, paramKey, paramVal string) (string, error) {
	u, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", err
	}
	q.Add(paramKey, paramVal)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
