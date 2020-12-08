package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/rester"
)

// HTTP formats
const (
	MimeTypeBlank          = ""
	MimeTypeHead           = "head"
	MimeTypeJSON           = "application/json"
	MimeTypeMultipartForm  = "multipart/form-data"
	MimeTypeURLEncodedForm = "application/x-www-form-urlencoded"
)

// Client represents a HTTP REST API client requesting a configured endpoint.
type Client struct {
	http.Client

	// the authentication layer
	authenticator authenticator

	// host & port of targeted endpoint
	url string

	// define usage of https
	secure bool
	// define if secure bool shall be considered to append the protocol to the URL
	ignoreProtocol bool
	// body response limit
	limit int64

	// overridable methods - mostly for formatting purpose
	get   func(context.Context, string, url.Values, interface{}) error
	post  func(context.Context, string, url.Values, interface{}, interface{}) error
	put   func(context.Context, string, url.Values, interface{}, interface{}) error
	patch func(context.Context, string, interface{}) error
}

// NewClient is HTTP Client constructor
func NewClient(url string, secure bool, options ...func(*Client)) *Client {
	cli := &Client{
		Client:         http.Client{},
		url:            url,
		secure:         secure,
		ignoreProtocol: false,
		limit:          1024 * 1024,
	}

	// by default
	// we consider the client is based on JSON formatting
	SetFormat(MimeTypeJSON)(cli)

	// run all potential options to set up the client
	for _, option := range options {
		option(cli)
	}
	return cli
}

// SetAuthenticator sets the optional authenticator on the http client number of retries
func SetAuthenticator(authenticator authenticator) func(*Client) {
	return func(c *Client) {
		c.authenticator = authenticator
	}
}

// SetFormat of the client by override some HTTP verb corresponding methods
func SetFormat(format string) func(*Client) {
	return func(c *Client) {
		switch format {
		case MimeTypeURLEncodedForm:
			c.get, c.post, c.put, c.patch = c.urlFormGet, c.urlFormPost, c.urlFormPut, c.urlFormPatch
		case MimeTypeMultipartForm:
			c.get, c.post, c.put, c.patch = c.multipartGet, c.multipartPost, c.multipartPut, c.multipartPatch
		default:
			c.get, c.post, c.put, c.patch = c.jsonGet, c.jsonPost, c.jsonPut, c.jsonPatch
		}
	}
}

// IgnoreProtocol by setting corresponding boolean to true
func IgnoreProtocol() func(*Client) {
	return func(c *Client) {
		c.ignoreProtocol = true
	}
}

// IgnoreInsecureHTTPS certificates
// shall never be used in production
func IgnoreInsecureHTTPS() func(*Client) {
	return func(c *Client) {
		c.Client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
}

// Head ...
// Fixed HTTP method
// Head entity using route as base url then params as query parameters
// A head request is agnostic from the format since no body input/output format is considered
// the output is built based on http headers
func (r *Client) Head(ctx context.Context, route string, params url.Values, output map[string][]string) error {
	return r.Perform(ctx, "HEAD", route, params, nil, output, MimeTypeHead)
}

// Delete ...
// Fixed HTTP method
// Delete an entity using route as base url then params as query parameters
// A delete request is agnostic from the format since no body input/output is considered
func (r *Client) Delete(ctx context.Context, route string, params url.Values) error {
	return r.Perform(ctx, "DELETE", route, params, nil, nil, MimeTypeBlank)
}

// Post ...
// Overridable HTTP method
// post an entity using route as base url then params as query parameters
func (r *Client) Post(ctx context.Context, route string, params url.Values, input interface{}, output interface{}) error {
	return r.post(ctx, route, params, input, output)
}

// Get ...
// Overridable HTTP method
// Get an entity using route as base url then params as query parameters
func (r *Client) Get(ctx context.Context, route string, params url.Values, output interface{}) error {
	return r.get(ctx, route, params, output)
}

// Put ...
// Overridable HTTP method
// put an entity using route as base url then params as query parameters
func (r *Client) Put(ctx context.Context, route string, params url.Values, input interface{}, output interface{}) error {
	return r.put(ctx, route, params, input, output)
}

// Patch ...
// Overridable HTTP method
// patch an entity using route as base url
func (r *Client) Patch(ctx context.Context, route string, input interface{}) error {
	return r.patch(ctx, route, input)
}

// Perform ...
func (r *Client) Perform(
	ctx context.Context,
	verb string,
	route string,
	params url.Values,
	input interface{},
	output interface{},
	format string,
) error {
	// 1. build URL, request, and use optional input to fill body
	req, err := http.NewRequest(verb, r.buildURL(r.secure, r.url, route, params), nil)
	if err != nil {
		return merror.Transform(err).Describe("could not create request")
	}
	if input != nil {
		var data []byte
		switch format {
		case MimeTypeJSON:
			data, err = json.Marshal(input)
			if err != nil {
				return merror.Transform(err).Describe("could not encode body")
			}
		case MimeTypeURLEncodedForm:
			params := input.(url.Values)
			data = []byte(params.Encode())
		default: // handle special MultipartForm mime type
			if strings.HasPrefix(format, MimeTypeMultipartForm) {
				buffer, ok := input.(*bytes.Buffer)
				if !ok {
					return merror.Internal().Describe("expecting input as a bytes.Buffer pointer")
				}
				data = buffer.Bytes()
			}
		}
		req.Header.Set("Content-Type", format)
		req.ContentLength = int64(len(data))
		req.Body = ioutil.NopCloser(bytes.NewReader(data))
	}

	// set authentication layer
	r.authenticator.Set(ctx, req)

	// set potential headers
	val := ctx.Value(rester.HeadersContextKey)
	if val != nil {
		headers := val.(map[string]string)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// 2. Perform request
	resp, err := r.Do(req)
	if err != nil {
		return merror.Transform(err).Describe("could perform request")
	}

	// Head format is a special case where we want to retrieve headers instead of body
	if format == MimeTypeHead {
		return handleHeaders(resp, output, r.limit)
	}
	return handleJSON(resp, output, r.limit)
}

func (r *Client) buildURL(secure bool, url string, route string, params url.Values) string {
	// configure protocol security
	if !r.ignoreProtocol {
		protocol := "http"
		if secure {
			protocol = "https"
		}
		url = fmt.Sprintf("%s://%s", protocol, url)
	}

	// build query string
	query := params.Encode()
	if len(query) > 0 {
		query = fmt.Sprintf("?%s", query)
	}

	return fmt.Sprintf("%s%s%s", url, route, query)
}
