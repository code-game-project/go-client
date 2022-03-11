package cg

import (
	"fmt"

	"github.com/gorilla/websocket"
)

// Connection represents the connection with a CodeGame server and handles events.
type Connection struct {
	gameId   string
	username string
	wsConn   *websocket.Conn
}

// Connect opens a new websocket connection with the CodeGame server listening at wsURL and returns a new Connection struct.
func Connect(wsURL, username string) (*Connection, error) {
	connection := &Connection{
		username: username,
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create websocket connection: %w", err)
	}
	connection.wsConn = wsConn

	// TODO set close handler

	return connection, nil
}

// Listen starts listening for events and triggers registered event listeners.
// Returns on close of connection.
func (c *Connection) Listen() error {
	for {
		msgType, msg, err := c.wsConn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err) {
				// TODO log error
				return err
			}
			break
		}
		if msgType != websocket.TextMessage {
			// TODO log error
			continue
		}

		// TODO decode event and call event listeners
		_ = msg
	}
	return nil
}
