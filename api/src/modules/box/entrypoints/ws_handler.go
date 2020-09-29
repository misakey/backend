package entrypoints

import (
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/application"
)

type WebsocketHandler struct {
	boxService     *application.BoxApplication
	allowedOrigins []string
}

func NewWebsocketHandler(allowedOrigins []string, boxService *application.BoxApplication) WebsocketHandler {
	return WebsocketHandler{
		allowedOrigins: allowedOrigins,
		boxService:     boxService,
	}
}
