package http

import (
	"context"
	"net/http"
)

type authenticator interface {
	Set(context.Context, *http.Request)
}
