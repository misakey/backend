package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

// output as structure to decode received application/json entity
func (r *Client) jsonGet(ctx context.Context, route string, params url.Values, output interface{}) error {
	return r.Perform(ctx, "GET", route, params, nil, output, MimeTypeBlank)
}

// input as structure to encode into application/json as a body
// output as structure to decode as an application/json entity
func (r *Client) jsonPost(ctx context.Context, route string, params url.Values, input interface{}, output interface{}) error {
	return r.Perform(ctx, "POST", route, params, input, output, MimeTypeJSON)
}

// input as structure to encode into application/json as a body
// output as structure to decode as an  application/json entity
func (r *Client) jsonPut(ctx context.Context, route string, params url.Values, input interface{}, output interface{}) error {
	return r.Perform(ctx, "PUT", route, params, input, output, MimeTypeJSON)
}

// input as structure to encode into application/json as a body
// no output expected
func (r *Client) jsonPatch(ctx context.Context, route string, input interface{}) error {
	return r.Perform(ctx, "PATCH", route, nil, input, nil, MimeTypeJSON)
}

// handleJSON takes a http response and try to unmarshal it as JSON into given output interface
func handleJSON(resp *http.Response, output interface{}, limit int64) error {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return merror.Transform(err).Describe("could not read response body")
	}

	// we consider an error occured below code 200 and above code 400
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return merror.TransformHTTPCode(resp.StatusCode).Describe(string(data))
	}

	// if we were supposed to retrieve an output, we try to unmarshal it
	if output != nil {
		err = json.Unmarshal(data, output)
		if err != nil {
			desc := fmt.Sprintf("could not decode output: %v (%v)", err, strings.Replace(string(data), "\n", "", -1))
			return merror.Transform(err).Describe(desc)
		}
	}
	return nil
}
