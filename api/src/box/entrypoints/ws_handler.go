package entrypoints

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/box/application"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/mwebsockets"
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

func (wh *WebsocketHandler) RedisListener(
	c echo.Context,
	wsName string,
	chName string,
	handler func(echo.Context, WebsocketHandler, []byte) error,
) error {
	ctx := c.Request().Context()
	ws, err := mwebsockets.NewWebsocket(c, wh.allowedOrigins, wsName)
	if err != nil {
		return err
	}

	go ws.Pump(c)

	sub := wh.boxService.RedConn.Subscribe(chName)
	// plan subscribe closing
	defer func(ctx context.Context, sub *redis.PubSub) {
		if err := sub.Close(); err != nil {
			logger.FromCtx(ctx).Error().Msgf("could not close sub: %v", err)
		}
	}(ctx, sub)

	// we instantiate this interface to wait for
	// the subscription to be active
	// see: https://godoc.org/github.com/go-redis/redis#Client.Subscribe
	iface, err := sub.Receive()
	if err != nil {
		return err
	}
	// Should be *Subscription, but others are possible if other actions have been
	// taken on sub since it was created.
	switch msg := iface.(type) {
	case *redis.Subscription:
		logger.FromCtx(ctx).Debug().Msg("Redis Subscription successful")
	case *redis.Message:
		ws.Send <- mwebsockets.WebsocketMessage{
			Msg: msg.Payload,
		}
	default:
		logger.FromCtx(ctx).Debug().Msgf("Unexpected redis notification type: %s", msg)
	}

	logger.FromCtx(ctx).Debug().Msg("Websocket loop started")
	subNotif := sub.Channel()
	for {
		select {
		case <-ws.EndPump:
			return nil
		case msg := <-ws.Handler:
			if err := handler(c, *wh, msg); err != nil {
				logger.FromCtx(ctx).Warn().Err(err).Msgf("%s: unable to handle msg", ws.ID)
			}
		case msg, ok := <-subNotif:
			if !ok {
				// here we close the websocket and send an error to the client
				logger.FromCtx(ctx).Error().Msg("cannot receive redis messages")
				_ = ws.SendCloseMessage()
				return merror.Internal().Describe("cannot receive redis messages")
			}
			ws.Send <- mwebsockets.WebsocketMessage{
				Msg: msg.Payload,
			}
		}
	}
}
