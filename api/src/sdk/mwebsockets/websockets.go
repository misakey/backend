/* This package is freely inspired by the whole websocket management
of the Mattermost server: https://github.com/mattermost/mattermost-server */

package mwebsockets

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

var (
	pingInterval     = 45 * time.Second
	pongWaitTime     = 120 * time.Second
	writeWaitTime    = 30 * time.Second
	closeGracePeriod = 1 * time.Second
	maxMessageSizekB = int64(8 * 1024)
	sendQueueSize    = 8
)

type Websocket struct {
	Websocket *websocket.Conn
	Send      chan WebsocketMessage
	Receive   chan ReceivedMessage
	Handler   chan []byte
	Interrupt chan struct{}
	EndPump   chan struct{}

	ID string
}

type WebsocketMessage struct {
	Msg string
}

type ReceivedMessage struct {
	msg []byte
	err error
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
		Websocket: wsConn,
		Send:      make(chan WebsocketMessage, sendQueueSize),
		Receive:   make(chan ReceivedMessage),
		Handler:   make(chan []byte),
		EndPump:   make(chan struct{}),
		Interrupt: make(chan struct{}),
		ID:        "ws_" + id,
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
		logger.FromCtx(eCtx.Request().Context()).Error().Err(err).Msgf("%s: read pump error", ws.ID)
	}
	close(ws.EndPump)
	logger.FromCtx(eCtx.Request().Context()).Debug().Msgf("%s: pump closed", ws.ID)
}

func (ws *Websocket) writePump() error {
	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case <-ws.EndPump:
			return nil
		case <-ticker.C:
			if err := ws.SendMessage(websocket.PingMessage, []byte{}); err != nil {
				return merror.Internal().Describef("%s: sending ping", ws.ID)
			}
		case msg, ok := <-ws.Send:
			if !ok {
				_ = ws.SendCloseMessage()
				return merror.Internal().Describef("%s: getting message", ws.ID)
			}

			toSend := []byte(msg.Msg)
			if err := ws.SendMessage(websocket.TextMessage, toSend); err != nil {
				return merror.Internal().Describef("%s: sending message", ws.ID)
			}
		}
	}
}

func (ws *Websocket) listener(eCtx echo.Context) {
	for {
		// NOTE: manage normal message when the client
		// communicates with the server
		_, msg, err := ws.Websocket.ReadMessage()
		ws.Receive <- ReceivedMessage{msg: msg, err: err}
		if err != nil {
			return
		}
	}
}

func (ws *Websocket) readPump(eCtx echo.Context) error {
	defer ws.Close()

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

	go ws.listener(eCtx)

	for {
		select {
		case recMsg := <-ws.Receive:
			if recMsg.err != nil {
				logger.FromCtx(eCtx.Request().Context()).Debug().Msgf("%s: error message: %s", ws.ID, recMsg.err.Error())
				return nil
			}
			logger.FromCtx(eCtx.Request().Context()).Debug().Msgf("%s: read message: %s", ws.ID, recMsg.msg)
			ws.Handler <- recMsg.msg
		case <-ws.Interrupt:
			// trying to gracefully close the socket
			_ = ws.SendCloseMessage()
			time.Sleep(closeGracePeriod)
			return nil
		}
	}
}

func (ws *Websocket) SendCloseMessage() error {
	return ws.SendMessage(websocket.CloseNormalClosure, []byte{})
}

func (ws *Websocket) SendMessage(kind int, msg []byte) error {
	_ = ws.Websocket.SetWriteDeadline(time.Now().Add(writeWaitTime))
	return ws.Websocket.WriteMessage(kind, msg)
}

func (ws *Websocket) Close() error {
	return ws.Websocket.Close()
}
