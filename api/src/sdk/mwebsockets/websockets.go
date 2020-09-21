/* This package is freely inspired by the whole websocket management
of the Mattermost server: https://github.com/mattermost/mattermost-server */

package mwebsockets

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

var (
	pingInterval     = 45 * time.Second
	pongWaitTime     = 120 * time.Second
	writeWaitTime    = 30 * time.Second
	maxMessageSizekB = int64(8 * 1024)
	sendQueueSize    = 8
)

type Websocket struct {
	Websocket    *websocket.Conn
	Send         chan WebsocketMessage
	EndWritePump chan struct{}

	ID string
}

type WebsocketMessage struct {
	Msg string
}

func NewWebsocket(
	eCtx echo.Context,
	allowedOrigins []string,
	id string,
) (*Websocket, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(req *http.Request) bool {
			origin := req.Header.Get("Origin")
			if origin == "" {
				return true
			}

			for _, allowed := range allowedOrigins {
				if allowed == origin || allowed == "*" {
					return true
				}
			}
			return false
		},
	}

	wsConn, err := upgrader.Upgrade(eCtx.Response(), eCtx.Request(), nil)
	if err != nil {
		return nil, err
	}

	ws := &Websocket{
		Websocket:    wsConn,
		Send:         make(chan WebsocketMessage, sendQueueSize),
		EndWritePump: make(chan struct{}),
		ID:           "ws_" + id,
	}

	return ws, nil
}

func (ws *Websocket) Pump(eCtx echo.Context) {
	logger.FromCtx(eCtx.Request().Context()).Debug().Msgf("%s: init pump", ws.ID)
	// writePump routine
	go func() {
		if err := ws.writePump(); err != nil {
			logger.FromCtx(eCtx.Request().Context()).Error().Err(err).Msgf("%s", ws.ID)
		}
	}()

	// readPump routine
	if err := ws.readPump(eCtx); err != nil {
		logger.FromCtx(eCtx.Request().Context()).Error().Err(err).Msgf("%s", ws.ID)
	}

	close(ws.EndWritePump)
	logger.FromCtx(eCtx.Request().Context()).Debug().Msgf("%s: pump closed", ws.ID)
}

func (ws *Websocket) writePump() error {
	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case msg, ok := <-ws.Send:
			if !ok {
				_ = ws.SendCloseMessage()
				return merror.Internal().Describef("%s: getting message", ws.ID)
			}

			toSend := []byte(msg.Msg)
			if err := ws.SendMessage(websocket.TextMessage, toSend); err != nil {
				return merror.Internal().Describef("%s: sending message", ws.ID)
			}
		case <-ticker.C:
			if err := ws.SendMessage(websocket.PingMessage, []byte{}); err != nil {
				return merror.Internal().Describef("%s: sending ping", ws.ID)
			}
		case <-ws.EndWritePump:
			return nil
		}
	}
}

func (ws *Websocket) readPump(eCtx echo.Context) error {
	ws.Websocket.SetReadLimit(maxMessageSizekB)

	_ = ws.Websocket.SetReadDeadline(time.Now().Add(pongWaitTime))
	ws.Websocket.SetPongHandler(func(string) error {
		_ = ws.Websocket.SetReadDeadline(time.Now().Add(pongWaitTime))
		logger.FromCtx(eCtx.Request().Context()).Debug().Msgf("%s: pong received", ws.ID)
		// TODO: handle this when we know how the browser reacts
		// if we return an error, the connection is closed
		// which may not be what we want if the browser does not
		// send Pongs on inactive tab for example
		return nil
	})

	for {
		_, _, err := ws.Websocket.ReadMessage()
		if err != nil {
			logger.FromCtx(eCtx.Request().Context()).Debug().Msgf("%s: read message: %s", ws.ID, err.Error())
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				return err
			}
			return nil
		}
	}
}

func (ws *Websocket) SendCloseMessage() error {
	return ws.SendMessage(websocket.CloseMessage, []byte{})
}

func (ws *Websocket) SendMessage(kind int, msg []byte) error {
	_ = ws.Websocket.SetWriteDeadline(time.Now().Add(writeWaitTime))
	return ws.Websocket.WriteMessage(kind, msg)
}

func (ws *Websocket) Close() error {
	close(ws.Send)
	return ws.Websocket.Close()
}
