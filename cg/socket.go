package cg

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrEncodeFailed       = errors.New("failed to encode json object")
	ErrDecodeFailed       = errors.New("failed to decode event")
)

// Socket represents the connection with a CodeGame server and handles events.
type Socket struct {
	state          State
	wsConn         *websocket.Conn
	eventListeners map[EventName]map[CallbackId]OnEventCallback
	usernameCache  map[string]string
}

// NewSocket opens a new websocket connection with the CodeGame server listening at wsURL and returns a new Connection struct.
func NewSocket(game string, wsURL string) (*Socket, error) {
	socket := &Socket{
		eventListeners: make(map[EventName]map[CallbackId]OnEventCallback),
		usernameCache:  make(map[string]string),
		state: State{
			Name: game,
		},
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create websocket connection: %w", err)
	}
	socket.wsConn = wsConn

	socket.On(EventJoinedGame, func(origin string, target EventTarget, event Event) {
		var data EventJoinedGameData
		event.UnmarshalData(&data)
		socket.cacheUser(origin, data.Username)
	})
	socket.On(EventLeftGame, func(origin string, target EventTarget, event Event) {
		var data EventLeftGameData
		event.UnmarshalData(&data)
		socket.uncacheUser(origin)
	})
	socket.On(EventGameInfo, func(origin string, target EventTarget, event Event) {
		var data EventGameInfoData
		event.UnmarshalData(&data)
		for id, name := range data.Players {
			socket.cacheUser(id, name)
		}
	})

	return socket, nil
}

// Create sends a create_game event to the server and returns the gameId on success.
func (s *Socket) Create() (string, error) {
	res, err := s.sendEventAndWaitForResponse(EventCreateGame, EventCreateGameData{}, EventCreatedGame)
	if err != nil {
		return "", err
	}

	var data EventCreatedGameData
	err = res[0].Event.UnmarshalData(&data)
	return data.GameId, err
}

// Join sends a join_game event to the server and returns a State object once it receives a joined_game and a player_secret event.
func (s *Socket) Join(gameId, username string) (State, error) {
	res, err := s.sendEventAndWaitForResponse(EventJoinGame, EventJoinGameData{
		GameId:   gameId,
		Username: username,
	}, EventJoinedGame, EventPlayerSecret)
	if err != nil {
		return State{}, err
	}

	var joinedData EventJoinedGameData
	err = res[0].Event.UnmarshalData(&joinedData)
	if err != nil {
		return State{}, err
	}

	var playerSecretData EventPlayerSecretData
	err = res[1].Event.UnmarshalData(&playerSecretData)
	if err != nil {
		return State{}, err
	}

	s.state.GameId = gameId
	s.state.PlayerId = res[0].Origin
	s.state.PlayerSecret = playerSecretData.Secret

	return s.state, nil
}

// Connect sends a connect_game event to the server and returns once it receives a connected_game event.
func (s *Socket) Connect(gameId, playerId, secret string) error {
	_, err := s.sendEventAndWaitForResponse(EventConnect, EventConnectData{
		GameId:   gameId,
		PlayerId: playerId,
		Secret:   secret,
	}, EventConnected)
	return err
}

// sendEventAndWaitForResponse sends event with eventData and waits until it receives expectedReponse.
// sendEventAndWaitForResponse returns the response event.
// Registered event listeners will be triggered.
func (s *Socket) sendEventAndWaitForResponse(event EventName, eventData any, expectedReponse ...EventName) ([]eventWrapper, error) {
	s.Emit(event, eventData)

	events := make([]eventWrapper, len(expectedReponse))

	for {
		wrapper, err := s.receiveEvent()
		if err != nil {
			return []eventWrapper{}, err
		}
		s.triggerEventListeners(wrapper.Origin, wrapper.Target, wrapper.Event)

		for i, expected := range expectedReponse {
			if wrapper.Event.Name == expected {
				events[i] = wrapper
				break
			}
		}

		done := true
		for i, expected := range expectedReponse {
			if events[i].Event.Name != expected {
				done = false
				break
			}
		}
		if done {
			return events, nil
		}

		if wrapper.Event.Name == EventError {
			var data EventErrorData
			wrapper.Event.UnmarshalData(&data)
			return []eventWrapper{}, errors.New(data.Reason)
		}
	}
}

// Listen starts listening for events and triggers registered event listeners.
// Returns on close or error.
func (s *Socket) Listen() error {
	for {
		wrapper, err := s.receiveEvent()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseGoingAway) {
				return nil
			} else {
				return err
			}
		}
		s.triggerEventListeners(wrapper.Origin, wrapper.Target, wrapper.Event)
	}
}

func (s *Socket) receiveEvent() (eventWrapper, error) {
	msgType, msg, err := s.wsConn.ReadMessage()
	if err != nil {
		return eventWrapper{}, err
	}
	if msgType != websocket.TextMessage {
		return eventWrapper{}, ErrInvalidMessageType
	}

	var wrapper eventWrapper
	err = json.Unmarshal(msg, &wrapper)
	if err != nil {
		return eventWrapper{}, ErrDecodeFailed
	}
	if wrapper.Event.Name == "" {
		return eventWrapper{}, ErrDecodeFailed
	}

	return wrapper, nil
}

// On registers a callback that is triggered when event is received.
func (s *Socket) On(event EventName, callback OnEventCallback) CallbackId {
	if s.eventListeners[event] == nil {
		s.eventListeners[event] = make(map[CallbackId]OnEventCallback)
	}

	id := CallbackId(uuid.New())

	s.eventListeners[event][id] = callback

	return id
}

// OnOnce registers a callback that is triggered only the first time event is received.
func (s *Socket) OnOnce(event EventName, callback OnEventCallback) CallbackId {
	if s.eventListeners[event] == nil {
		s.eventListeners[event] = make(map[CallbackId]OnEventCallback)
	}

	id := CallbackId(uuid.New())

	s.eventListeners[event][id] = func(origin string, target EventTarget, event Event) {
		callback(origin, target, event)
		s.RemoveCallback(id)
	}

	return id
}

// RemoveCallback deletes the callback with the specified id.
func (s *Socket) RemoveCallback(id CallbackId) {
	for _, callbacks := range s.eventListeners {
		delete(callbacks, id)
	}
}

// Emit sends a new event to the server.
func (s *Socket) Emit(eventName EventName, eventData interface{}) error {
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

	s.wsConn.WriteMessage(websocket.TextMessage, jsonData)
	return nil
}

// Leave sends a leave_game event to the server and clears all non-standard event listeners.
func (s *Socket) Leave() error {
	for key := range s.eventListeners {
		if !IsStandardEvent(key) {
			delete(s.eventListeners, key)
		}
	}

	for key := range s.usernameCache {
		delete(s.usernameCache, key)
	}

	s.state.Remove()

	return s.Emit(EventLeaveGame, EventLeaveGameData{})
}

// Close closes the underlying websocket connection.
func (s *Socket) Close() error {
	s.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(5*time.Second))
	return s.wsConn.Close()
}

// Returns the username associated with playerId.
func (s *Socket) GetUser(playerId string) string {
	return s.usernameCache[playerId]
}

func (s *Socket) triggerEventListeners(origin string, target EventTarget, event Event) {
	if s.eventListeners[event.Name] != nil {
		for _, cb := range s.eventListeners[event.Name] {
			cb(origin, target, event)
		}
	}
}

func (s *Socket) cacheUser(playerId, username string) {
	s.usernameCache[playerId] = username
}

func (s *Socket) uncacheUser(playerId string) {
	delete(s.usernameCache, playerId)
}
