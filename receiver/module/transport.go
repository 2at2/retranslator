package module

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/mono83/slf/wd"
	"strconv"
	"time"
)

// WebsocketTransport is wrapper over HTTP connection and is used to forward messages to frontend
type WebsocketTransport struct {
	c *websocket.Conn
}

// IsAlive returns true if socket connection still alive
func (w *WebsocketTransport) IsAlive() bool {
	return w.c != nil
}

func (w *WebsocketTransport) Ping() error {
	if !w.IsAlive() {
		return errors.New("socket is closed")
	}

	w.c.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))

	if err := w.c.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		log.Error("Failed ping - :err", wd.ErrParam(err))
		w.Close()
		return err
	}

	return nil
}

// Send sends packet
func (w *WebsocketTransport) Write(msg []byte) error {
	if !w.IsAlive() {
		return errors.New("socket is closed")
	}

	w.c.SetWriteDeadline(time.Now().Add(5 * time.Second)) // TODO

	if err := w.c.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Error("Failed write - :err", wd.ErrParam(err))
		w.Close()
		return err
	}

	return nil
}

// Read returns message
func (w *WebsocketTransport) Read() ([]byte, error) {
	if !w.IsAlive() {
		return nil, errors.New("socket is closed")
	}

	w.c.SetReadDeadline(time.Now().Add(10 * time.Second)) // TODO

	messageType, messageBody, err := w.c.ReadMessage()

	if err != nil {
		log.Error("Failed read - :err", wd.ErrParam(err))
		w.Close()
		return nil, err
	}

	switch messageType {
	case websocket.TextMessage:
		return messageBody, nil
	case websocket.CloseMessage:
		log.Warning("Closing socket")
		w.Close()
		return nil, errors.New("socket is closed")
	}

	return nil, errors.New("unknown message type " + strconv.Itoa(messageType))
}

func (w *WebsocketTransport) Close() error {
	if w.IsAlive() {
		w.c.SetWriteDeadline(time.Now().Add(1 * time.Second)) // TODO

		if err := w.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
			log.Error("Failed to close gracefully - :err", wd.ErrParam(err))
		}

		w.c.Close()
		w.c = nil
	}

	return nil
}
