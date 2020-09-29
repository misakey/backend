package entrypoints

import (
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/application"
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
	interruptChan string,
	handler func(echo.Context, WebsocketHandler, []byte) error,
) error {
	ws, err := mwebsockets.NewWebsocket(c, wh.allowedOrigins, wsName)
	if err != nil {
		return err
	}

	go ws.Pump(c)
	if interruptChan != "" {
		go listenInterrupt(c, ws, wh.boxService.RedConn, interruptChan)
	}

	sub := wh.boxService.RedConn.Subscribe(chName)

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
		logger.FromCtx(c.Request().Context()).Debug().Msg("Redis Subscription successful")
	case *redis.Message:
		ws.Send <- mwebsockets.WebsocketMessage{
			Msg: msg.Payload,
		}
	default:
		logger.FromCtx(c.Request().Context()).Debug().Msgf("Unexpected redis notification type: %s", msg)
	}

	ch := sub.Channel()

	logger.FromCtx(c.Request().Context()).Debug().Msg("Websocket loop started")
	for {
		select {
		case <-ws.EndPump:
			return nil
		case msg := <-ws.Handler:
			if err := handler(c, *wh, msg); err != nil {
				logger.FromCtx(c.Request().Context()).Warn().Err(err).Msgf("%s: unable to handle msg", ws.ID)
			}
		case msg, ok := <-ch:
			if !ok {
				// here we close the websocket and send an error to the client
				logger.FromCtx(c.Request().Context()).Error().Msg("cannot receive redis messages")
				_ = ws.SendCloseMessage()
				return merror.Internal().Describe("cannot receive redis messages")
			}
			ws.Send <- mwebsockets.WebsocketMessage{
				Msg: msg.Payload,
			}
		}
	}
}

func listenInterrupt(c echo.Context, ws *mwebsockets.Websocket, redConn *redis.Client, interruptChan string) {
	sub := redConn.Subscribe(interruptChan)
	// we instantiate this interface to wait for
	// the subscription to be active
	// see: https://godoc.org/github.com/go-redis/redis#Client.Subscribe
	iface, err := sub.Receive()
	if err != nil {
		return
	}
	// Should be *Subscription, but others are possible if other actions have been
	// taken on sub since it was created.
	switch msg := iface.(type) {
	case *redis.Subscription:
		logger.FromCtx(c.Request().Context()).Debug().Msgf("%s: redis Subscription successful", ws.ID)
	case *redis.Message:
		ws.Send <- mwebsockets.WebsocketMessage{
			Msg: msg.Payload,
		}
	default:
		logger.FromCtx(c.Request().Context()).Debug().Msgf("%s: unexpected redis notification type: %s", msg, ws.ID)
	}

	ch := sub.Channel()

	logger.FromCtx(c.Request().Context()).Debug().Msgf("%s: websocket loop started", ws.ID)
	for {
		select {
		case <-ws.EndPump:
			return
		case msg, ok := <-ch:
			if !ok {
				logger.FromCtx(c.Request().Context()).Error().Msgf("%s: cannot receive redis messages", ws.ID)
			}
			if msg.Payload == "stop" {
				logger.FromCtx(c.Request().Context()).Error().Msgf("%s: interrupting websocket", ws.ID)
				close(ws.Interrupt)
			} else {
				logger.
					FromCtx(c.Request().Context()).
					Error().
					Msgf("%s: unexpected message in interrupt chan: %s", msg.String(), ws.ID)
			}
		}
	}
}
