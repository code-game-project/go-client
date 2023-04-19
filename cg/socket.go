package cg

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrEncodeFailed       = errors.New("failed to encode json object")
	ErrDecodeFailed       = errors.New("failed to decode event")
	ErrClosed             = errors.New("connection closed")
)

// Socket represents the connection with a CodeGame server and handles events.
type Socket struct {
	gameURL        string
	tls            bool
	wsConn         *websocket.Conn
	eventListeners map[EventName]map[CallbackID]EventCallback
	usernameCache  map[string]string

	gameID   string
	playerID string

	running   bool
	eventChan chan Event
	err       error

	nextCallbackID CallbackID
}

func Connect(gameURL, gameID, playerID, playerSecret string) (*Socket, error) {
	gameURL = trimURL(gameURL)
	socket := &Socket{
		gameURL:        gameURL,
		tls:            isTLS(gameURL),
		eventListeners: make(map[EventName]map[CallbackID]EventCallback),
		usernameCache:  make(map[string]string),
		eventChan:      make(chan Event, 10),
		gameID:         gameID,
		playerID:       playerID,
	}
	err := socket.connect(gameID, playerID, playerSecret)
	if err != nil {
		return nil, err
	}

	socket.startListenLoop()

	socket.usernameCache, err = socket.fetchPlayers(gameID)
	if err != nil {
		return nil, err
	}

	return socket, nil
}

func Spectate(gameURL, gameID string) error {
	gameURL = trimURL(gameURL)
	socket := &Socket{
		gameURL:        gameURL,
		tls:            isTLS(gameURL),
		eventListeners: make(map[EventName]map[CallbackID]EventCallback),
		usernameCache:  make(map[string]string),
		eventChan:      make(chan Event, 10),
		gameID:         gameID,
	}
	err := socket.spectate(gameID)
	if err != nil {
		return err
	}

	socket.startListenLoop()

	socket.usernameCache, err = socket.fetchPlayers(gameID)
	if err != nil {
		return err
	}

	return nil
}

// RunEventLoop starts listening for events and triggers registered event listeners.
// Returns on close or error.
func (s *Socket) RunEventLoop() error {
	for s.running {
		event, ok := <-s.eventChan
		if !ok {
			break
		}
		s.triggerEventListeners(event)
	}
	if s.err == ErrClosed {
		return nil
	}
	return s.err
}

// NextEvent returns the next event in the queue or ok = false if there is none.
// Registered event listeners will be triggered.
func (s *Socket) NextEvent() (Event, bool, error) {
	select {
	case event, ok := <-s.eventChan:
		if ok {
			s.triggerEventListeners(event)
			return event, true, nil
		} else {
			return Event{}, false, s.err
		}
	default:
		return Event{}, false, nil
	}
}

// On registers a callback that is triggered when the event is received.
func (s *Socket) On(event EventName, callback EventCallback) CallbackID {
	if s.eventListeners[event] == nil {
		s.eventListeners[event] = make(map[CallbackID]EventCallback)
	}

	id := s.nextCallbackID
	s.nextCallbackID++

	s.eventListeners[event][id] = callback

	return id
}

// Once registers a callback that is triggered only the first time the event is received.
func (s *Socket) Once(event EventName, callback EventCallback) CallbackID {
	if s.eventListeners[event] == nil {
		s.eventListeners[event] = make(map[CallbackID]EventCallback)
	}

	id := s.nextCallbackID
	s.nextCallbackID++

	s.eventListeners[event][id] = func(event Event) {
		callback(event)
		s.RemoveCallback(id)
	}

	return id
}

// RemoveCallback deletes the callback with the specified id.
func (s *Socket) RemoveCallback(id CallbackID) {
	for _, callbacks := range s.eventListeners {
		delete(callbacks, id)
	}
}

// Send sends a new command to the server.
// Send panics if the socket is not connected to a player.
func (s *Socket) Send(name CommandName, data any) error {
	if s.playerID == "" {
		panic("cannot send commands as a spectator")
	}

	cmd := Command{
		Name: name,
	}

	if data == nil {
		data = struct{}{}
	}

	err := cmd.marshalData(data)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	s.wsConn.WriteMessage(websocket.TextMessage, jsonData)
	return nil
}

// Close closes the underlying websocket connection.
func (s *Socket) Close() error {
	s.running = false
	s.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(5*time.Second))
	return s.wsConn.Close()
}

// Username returns the username associated with playerId.
func (s *Socket) Username(playerID string) string {
	if username, ok := s.usernameCache[playerID]; ok {
		return username
	}

	username, err := s.fetchUsername(s.gameID, playerID)
	if err == nil {
		s.usernameCache[playerID] = username
	}
	return username
}

func (s *Socket) GameURL() string {
	return s.gameURL
}

func (s *Socket) GameID() string {
	return s.gameID
}

func (s *Socket) PlayerID() string {
	return s.playerID
}

func (s *Socket) IsSpectating() bool {
	return s.playerID == ""
}

func (s *Socket) startListenLoop() {
	s.running = true
	go func() {
		for s.running {
			event, err := s.receiveEvent()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseGoingAway) {
					s.err = ErrClosed
				} else {
					s.err = err
				}
				s.running = false
				close(s.eventChan)
				continue
			}
			s.eventChan <- event
		}
	}()
}

func (s *Socket) receiveEvent() (Event, error) {
	msgType, msg, err := s.wsConn.ReadMessage()
	if err != nil {
		return Event{}, err
	}
	if msgType != websocket.TextMessage {
		return Event{}, ErrInvalidMessageType
	}

	var event Event
	err = json.Unmarshal(msg, &event)
	if err != nil {
		return Event{}, ErrDecodeFailed
	}
	if event.Name == "" {
		return Event{}, ErrDecodeFailed
	}

	return event, nil
}

func (s *Socket) triggerEventListeners(event Event) {
	listeners := s.eventListeners[event.Name]
	for _, cb := range listeners {
		cb(event)
	}
}
