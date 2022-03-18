package cg

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// Connection represents the connection with a CodeGame server and handles events.
type Connection struct {
	gameId         string
	username       string
	wsConn         *websocket.Conn
	eventListeners map[EventName][]OnEventCallback
}

// Connect opens a new websocket connection with the CodeGame server listening at wsURL and returns a new Connection struct.
func Connect(wsURL, username string) (*Connection, error) {
	connection := &Connection{
		username:       username,
		eventListeners: make(map[EventName][]OnEventCallback),
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create websocket connection: %w", err)
	}
	connection.wsConn = wsConn

	return connection, nil
}

// Listen starts listening for events and triggers registered event listeners.
// Returns on close or error.
func (c *Connection) Listen() error {
	for {
		msgType, msg, err := c.wsConn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseGoingAway) {
				return err
			}
			break
		}
		if msgType != websocket.TextMessage {
			c.error(fmt.Sprintf("received invalid message type"))
			continue
		}

		type eventWrapper struct {
			Target EventTarget `json:"target"`
			Origin string      `json:"origin"`
			Event  Event       `json:"event"`
		}
		var wrapper eventWrapper
		err = json.Unmarshal(msg, &wrapper)
		if err != nil {
			c.error(fmt.Sprintf("failed to decode event: %s", err))
			continue
		}
		if wrapper.Event.Name == "" {
			c.error(fmt.Sprintf("failed to decode event: empty event name field"))
			continue
		}

		c.triggerEventListeners(wrapper.Origin, wrapper.Target, wrapper.Event)
	}
	return nil
}

// On registers a callback that is triggered once event is received.
func (c *Connection) On(event EventName, callback OnEventCallback) {
	if c.eventListeners[event] == nil {
		c.eventListeners[event] = make([]OnEventCallback, 0, 1)
	}

	c.eventListeners[event] = append(c.eventListeners[event], callback)
}

// Emit sends a new event to the server.
func (c *Connection) Emit(eventName EventName, eventData interface{}) error {
	event := Event{
		Name: eventName,
	}
	err := event.marshalData(eventData)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	c.wsConn.WriteMessage(websocket.TextMessage, jsonData)
	return nil
}

// Close closes the underlying websocket connection.
func (c *Connection) Close() error {
	c.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(5*time.Second))
	return c.wsConn.Close()
}

func (c *Connection) triggerEventListeners(origin string, target EventTarget, event Event) {
	if c.eventListeners[event.Name] != nil {
		for _, cb := range c.eventListeners[event.Name] {
			cb(origin, target, event)
		}
	}
}

func (c *Connection) error(reason string) {
	errorEvent := Event{
		Name: ErrorEvent,
	}
	err := errorEvent.marshalData(ErrorEventData{
		Reason: reason,
	})
	if err == nil {
		c.triggerEventListeners(EventOriginSelf, EventTarget{Type: EventTargetTypeSelf}, errorEvent)
	}
}
