/* This package is freely inspired by the whole websocket management
of the Mattermost server: https://github.com/mattermost/mattermost-server */

package mwebsockets

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/msk-sdk-go/logger"
	"gitlab.misakey.dev/misakey/msk-sdk-go/merror"
)

var (
	pingInterval     = 60 * time.Second
	pongWaitTime     = 100 * time.Second
	writeWaitTime    = 30 * time.Second
	maxMessageSizekB = int64(8 * 1024)
	sendQueueSize    = 8
)

type Websocket struct {
	Websocket    *websocket.Conn
	Send         chan WebsocketMessage
	EndWritePump chan struct{}
}

type WebsocketMessage struct {
	Msg string
}

func NewWebsocket(
	eCtx echo.Context,
	allowedOrigins []string,
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
	}

	return ws, nil
}

func (ws *Websocket) Pump(eCtx echo.Context) {
	// writePump routine
	go func() {
		if err := ws.writePump(); err != nil {
			logger.FromCtx(eCtx.Request().Context()).Error().Err(err)
		}
	}()

	// readPump routine
	if err := ws.readPump(); err != nil {
		logger.FromCtx(eCtx.Request().Context()).Error().Err(err)
	}

	close(ws.EndWritePump)
}

func (ws *Websocket) writePump() error {
	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case msg, ok := <-ws.Send:
			if !ok {
				_ = ws.SendCloseMessage()
				return merror.Internal().Describe("getting websocket message")
			}

			toSend := []byte(msg.Msg)
			if err := ws.SendMessage(websocket.TextMessage, toSend); err != nil {
				return merror.Internal().Describe("sending websocket message")
			}
		case <-ticker.C:
			if err := ws.SendMessage(websocket.PingMessage, []byte{}); err != nil {
				return merror.Internal().Describe("sending ping")
			}
		case <-ws.EndWritePump:
			return nil
		}
	}
}

func (ws *Websocket) readPump() error {
	ws.Websocket.SetReadLimit(maxMessageSizekB)
	_ = ws.Websocket.SetReadDeadline(time.Now().Add(pongWaitTime))
	ws.Websocket.SetPongHandler(func(string) error {
		_ = ws.Websocket.SetReadDeadline(time.Now().Add(pongWaitTime))
		// TODO: handle this when we know how the browser reacts
		// if we return an error, the connection is closed
		// which may not be what we want if the browser does not
		// send Pongs on inactive tab for example
		return nil
	})

	for {
		var req WebsocketRequest
		if err := ws.Websocket.ReadJSON(&req); err != nil {
			return err
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

// WebsocketRequest represents a request made to the server through a websocket.
type WebsocketRequest struct {
}

func (o *WebsocketRequest) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func WebSocketRequestFromJson(data io.Reader) *WebsocketRequest {
	var o *WebsocketRequest
	_ = json.NewDecoder(data).Decode(&o)
	return o
}
