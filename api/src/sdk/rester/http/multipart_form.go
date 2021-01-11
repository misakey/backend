package http

import (
	"bytes"
	"context"
	"net/url"
	"reflect"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
)

// MultipartRequest ...
type MultipartRequest struct {
	Body        *bytes.Buffer
	ContentType string
	Output      interface{}
}

// Get an entity using jsonGet method - no multipart/form-data decoding implemented
func (r *Client) multipartGet(ctx context.Context, route string, params url.Values, output interface{}) error {
	return r.jsonGet(ctx, route, params, output)
}

// input as structure to encode into multipart/form-data as a body
// output as structure to decode as an application/json entity  - no multipart/form-data form decoding implemented
func (r *Client) multipartPost(ctx context.Context, route string, _ url.Values, input interface{}, output interface{}) error {
	mb, ok := input.(*MultipartRequest)
	if !ok {
		return merr.Internal().Descf("input shall be a multipart request type, received %s", reflect.TypeOf(input))
	}
	return r.Perform(ctx, "POST", route, nil, mb.Body, mb.Output, mb.ContentType)
}

// input as structure to encode into as a multipart/form-data body
// output as structure to decode as an application/json entity - no multipart/form-data form decoding implemented
func (r *Client) multipartPut(ctx context.Context, route string, _ url.Values, input interface{}, output interface{}) error {
	mb, ok := input.(*MultipartRequest)
	if !ok {
		return merr.Internal().Descf("input shall be a multipart request type, received %s", reflect.TypeOf(input))
	}
	return r.Perform(ctx, "PUT", route, nil, mb.Body, mb.Output, mb.ContentType)
}

// input as structure to encode into as a multipart/form-data body
// no output expected
func (r *Client) multipartPatch(ctx context.Context, route string, input interface{}) error {
	mb, ok := input.(*MultipartRequest)
	if !ok {
		return merr.Internal().Descf("input shall be a multipart request type, received %s", reflect.TypeOf(input))
	}
	return r.Perform(ctx, "PATCH", route, nil, mb.Body, nil, mb.ContentType)
}
