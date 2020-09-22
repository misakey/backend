package rester

import (
	"context"
	"encoding/base64"
	"net/url"
)

const HeadersContextKey = "headers"

type Client interface {
	Get(ctx context.Context, route string, params url.Values, output interface{}) error
	Head(ctx context.Context, route string, params url.Values, output map[string][]string) error
	Post(ctx context.Context, route string, params url.Values, input interface{}, output interface{}) error
	Put(ctx context.Context, route string, params url.Values, input interface{}, output interface{}) error
	Patch(ctx context.Context, route string, input interface{}) error
	Delete(ctx context.Context, route string, params url.Values) error
}

func GetBasicAuthHeader(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
