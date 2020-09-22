package http

import (
	"context"
	"net/url"
)

// Get an entity using jsonGet method - no urlencoded form decoding implemented
func (r *Client) urlFormGet(ctx context.Context, route string, params url.Values, output interface{}) error {
	return r.jsonGet(ctx, route, params, output)
}

// input as structure to encode into application/x-www-form-urlencoded as a body
// output as structure to decode as an  application/json entity  - no urlencoded form decoding implemented
func (r *Client) urlFormPost(ctx context.Context, route string, params url.Values, input interface{}, output interface{}) error {
	return r.Perform(ctx, "POST", route, params, input, output, URLENCODED_FORM_MIME_TYPE)
}

// input as structure to encode into application/x-www-form-urlencoded as a body
// output as structure to decode as an  application/json entity - no urlencoded form decoding implemented
func (r *Client) urlFormPut(ctx context.Context, route string, params url.Values, input interface{}, output interface{}) error {
	return r.Perform(ctx, "PUT", route, params, input, output, URLENCODED_FORM_MIME_TYPE)
}

// input as structure to encode into application/x-www-form-urlencoded as a body
// no output expected
func (r *Client) urlFormPatch(ctx context.Context, route string, input interface{}) error {
	return r.Perform(ctx, "PATCH", route, nil, input, nil, URLENCODED_FORM_MIME_TYPE)
}
