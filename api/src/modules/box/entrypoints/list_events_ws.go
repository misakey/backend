package entrypoints

import (
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/ajwt"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/mwebsockets"

	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/events"
)

func (wh WebsocketHandler) ListEventsWS(c echo.Context) error {

	// bind and validate
	boxID := c.Param("id")

	if err := v.Validate(boxID, v.Required, is.UUIDv4); err != nil {
		return merror.BadRequest().From(merror.OriPath).Detail("id", merror.DVMalformed)
	}

	// check accesses
	acc := ajwt.GetAccesses(c.Request().Context())
	if acc == nil {
		return merror.Forbidden()
	}
	if err := events.MustMemberHaveAccess(
		c.Request().Context(),
		wh.boxService.DB,
		wh.boxService.RedConn,
		wh.boxService.Identities,
		boxID,
		acc.IdentityID,
	); err != nil {
		return err
	}

	ws, err := mwebsockets.NewWebsocket(c, wh.allowedOrigins, fmt.Sprintf("%s_%s", acc.IdentityID, boxID))
	if err != nil {
		return err
	}

	go ws.Pump(c)
	go listenInterrupt(c, ws, wh.boxService.RedConn, acc.IdentityID, boxID)

	sub := wh.boxService.RedConn.Subscribe(boxID + ":events")

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

// TODO(code structure): factorize this for use in other parts
func listenInterrupt(c echo.Context, ws *mwebsockets.Websocket, redConn *redis.Client, senderID, boxID string) {
	sub := redConn.Subscribe("interrupt:" + boxID + ":" + senderID)
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
				logger.FromCtx(c.Request().Context()).Debug().Msgf("%s: interrupting websocket", ws.ID)
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
