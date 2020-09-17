package entrypoints

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/mwebsockets"
	"gitlab.misakey.dev/misakey/msk-sdk-go/ajwt"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

func (wh WebsocketHandler) ListEventsWS(c echo.Context) error {

	// bind and validate
	boxID := c.Param("id")

	if err := v.Validate(boxID, v.Required, is.UUIDv4); err != nil {
		return err
	}

	// check accesses
	acc := ajwt.GetAccesses(c.Request().Context())
	if acc == nil {
		return merror.Forbidden()
	}
	if err := events.MustMemberHaveAccess(c.Request().Context(), wh.db, wh.identities, boxID, acc.IdentityID); err != nil {
		return err
	}

	ws, err := mwebsockets.NewWebsocket(c, wh.allowedOrigins)
	if err != nil {
		return err
	}
	defer ws.Close()

	go ws.Pump(c)

	sub := wh.redConn.Subscribe(boxID + ":events")

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
		case <-ws.EndWritePump:
			return nil
		}
	}
}
