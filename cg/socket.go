package cg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
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
	name           string
	domain         string
	ssl            bool
	session        session
	wsConn         *websocket.Conn
	eventListeners map[EventName]map[CallbackId]OnEventCallback
	usernameCache  map[string]string

	running          bool
	eventWrapperChan chan EventWrapper
	err              error
}

// NewSocket opens a new websocket connection with the CodeGame server listening at domain (e.g. my-game.io) and returns a new Connection struct.
func NewSocket(domain string) (*Socket, error) {
	if strings.HasPrefix(domain, "http://") {
		domain = strings.TrimPrefix(domain, "http://")
	} else if strings.HasPrefix(domain, "https://") {
		domain = strings.TrimPrefix(domain, "https://")
	} else if strings.HasPrefix(domain, "ws://") {
		domain = strings.TrimPrefix(domain, "ws://")
	} else if strings.HasPrefix(domain, "wss://") {
		domain = strings.TrimPrefix(domain, "wss://")
	}
	domain = strings.TrimSuffix(domain, "/")

	socket := &Socket{
		domain:           domain,
		eventListeners:   make(map[EventName]map[CallbackId]OnEventCallback),
		usernameCache:    make(map[string]string),
		eventWrapperChan: make(chan EventWrapper),
	}

	res, err := http.Get("https://" + domain)
	if err == nil {
		res.Body.Close()
		socket.ssl = true
	}

	type response struct {
		Name      string `json:"name"`
		CGVersion string `json:"cg_version"`
	}
	res, err = http.Get(socket.baseURL(false) + "/info")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer res.Body.Close()
	var body response
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode /info data: %w", err)
	}
	if body.Name == "" {
		return nil, fmt.Errorf("empty game name")
	}
	socket.name = body.Name

	if body.CGVersion != CGVersion {
		fmt.Fprintf(os.Stderr, "WARNING: CodeGame version mismatch. Server: v%s, client: v%s\n", body.CGVersion, CGVersion)
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(socket.baseURL(true)+"/ws", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create websocket connection: %w", err)
	}
	socket.wsConn = wsConn

	socket.On(EventNewPlayer, func(origin string, event Event) {
		var data EventNewPlayerData
		event.UnmarshalData(&data)
		socket.cacheUser(origin, data.Username)
	})
	socket.On(EventLeft, func(origin string, event Event) {
		var data EventLeftData
		event.UnmarshalData(&data)
		socket.uncacheUser(origin)
	})
	socket.On(EventInfo, func(origin string, event Event) {
		var data EventInfoData
		event.UnmarshalData(&data)
		for id, name := range data.Players {
			socket.cacheUser(id, name)
		}
	})

	return socket, nil
}

// Create creates a new game on the server and returns the id of the created game.
func (s *Socket) Create(public bool) (string, error) {
	type request struct {
		Public bool `json:"public"`
	}
	data, err := json.Marshal(request{
		Public: public,
	})
	if err != nil {
		return "", err
	}

	body := bytes.NewBuffer(data)
	resp, err := http.Post(s.baseURL(false)+"/games", "application/json", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type response struct {
		GameId string `json:"game_id"`
	}
	var r response
	err = json.Unmarshal(data, &r)
	return r.GameId, err
}

// Join sends a join_game event to the server and returns a Session object once it receives a joined_game and a player_secret event.
func (s *Socket) Join(gameId, username string) error {
	if s.session.Name != "" {
		return errors.New("already joined a game")
	}

	if username == "" {
		return errors.New("empty username")
	}

	res, err := s.sendEventAndWaitForResponse(EventJoin, EventJoinData{
		GameId:   gameId,
		Username: username,
	}, EventJoined)
	if err != nil {
		return err
	}

	var data EventJoinedData
	err = res.Event.UnmarshalData(&data)
	if err != nil {
		return err
	}

	s.session = newSession(s.name, username, gameId, res.Origin, data.Secret)
	err = s.session.save()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save session:", err)
	}

	s.startListenLoop()

	return nil
}

// Connect sends a connect_game event to the server and returns once it receives a connected_game event.
func (s *Socket) Connect(username string) error {
	var err error
	s.session, err = loadSession(s.name, username)
	if err != nil {
		return err
	}
	_, err = s.sendEventAndWaitForResponse(EventConnect, EventConnectData{
		GameId:   s.session.GameId,
		PlayerId: s.session.PlayerId,
		Secret:   s.session.PlayerSecret,
	}, EventConnected)
	if err != nil {
		s.session.remove()
		s.session = session{}
	} else {
		s.startListenLoop()
	}
	return err
}

// RunEventLoop starts listening for events and triggers registered event listeners.
// Returns on close or error.
func (s *Socket) RunEventLoop() error {
	for s.running {
		wrapper, ok := <-s.eventWrapperChan
		if !ok {
			break
		}
		s.triggerEventListeners(wrapper.Origin, wrapper.Event)
	}
	if s.err == ErrClosed {
		return nil
	}
	return s.err
}

// NextEvent returns the next event in the queue or ok = false if there is none.
// Registered event listeners will be triggered.
func (s *Socket) NextEvent() (EventWrapper, bool, error) {
	select {
	case wrapper, ok := <-s.eventWrapperChan:
		if ok {
			s.triggerEventListeners(wrapper.Origin, wrapper.Event)
			return wrapper, true, nil
		} else {
			return EventWrapper{}, false, s.err
		}
	default:
		return EventWrapper{}, false, nil
	}
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

	s.eventListeners[event][id] = func(origin string, event Event) {
		callback(origin, event)
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

// Send sends a new event to the server.
func (s *Socket) Send(eventName EventName, eventData any) error {
	event := Event{
		Name: eventName,
	}

	if eventData == nil {
		eventData = struct{}{}
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
// It also deletes the current session.
func (s *Socket) Leave() error {
	for key := range s.eventListeners {
		if !IsStandardEvent(key) {
			delete(s.eventListeners, key)
		}
	}

	for key := range s.usernameCache {
		delete(s.usernameCache, key)
	}

	s.session.remove()

	s.running = false

	return s.Send(EventLeave, nil)
}

// Close closes the underlying websocket connection.
func (s *Socket) Close() error {
	s.running = false
	s.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(5*time.Second))
	return s.wsConn.Close()
}

// Returns the username associated with playerId.
func (s *Socket) GetUser(playerId string) string {
	return s.usernameCache[playerId]
}

// sendEventAndWaitForResponse sends event with eventData and waits until it receives expectedReponse.
// sendEventAndWaitForResponse returns the response event.
// Registered event listeners will be triggered.
func (s *Socket) sendEventAndWaitForResponse(event EventName, eventData any, expectedReponse EventName) (EventWrapper, error) {
	s.Send(event, eventData)

	for {
		wrapper, err := s.receiveEvent()
		if err != nil {
			return EventWrapper{}, err
		}
		s.triggerEventListeners(wrapper.Origin, wrapper.Event)

		if wrapper.Event.Name == expectedReponse {
			return wrapper, nil
		}

		if wrapper.Event.Name == EventError {
			var data EventErrorData
			wrapper.Event.UnmarshalData(&data)
			return EventWrapper{}, errors.New(data.Reason)
		}
	}
}

func (s *Socket) startListenLoop() {
	s.running = true
	go func() {
		for s.running {
			wrapper, err := s.receiveEvent()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseGoingAway) {
					s.err = ErrClosed
				} else {
					s.err = err
				}
				s.running = false
				close(s.eventWrapperChan)
			} else {
				s.eventWrapperChan <- wrapper
			}
		}
	}()
}

func (s *Socket) receiveEvent() (EventWrapper, error) {
	msgType, msg, err := s.wsConn.ReadMessage()
	if err != nil {
		return EventWrapper{}, err
	}
	if msgType != websocket.TextMessage {
		return EventWrapper{}, ErrInvalidMessageType
	}

	var wrapper EventWrapper
	err = json.Unmarshal(msg, &wrapper)
	if err != nil {
		return EventWrapper{}, ErrDecodeFailed
	}
	if wrapper.Event.Name == "" {
		return EventWrapper{}, ErrDecodeFailed
	}

	return wrapper, nil
}

func (s *Socket) triggerEventListeners(origin string, event Event) {
	if s.eventListeners[event.Name] != nil {
		for _, cb := range s.eventListeners[event.Name] {
			cb(origin, event)
		}
	}
}

func (s *Socket) cacheUser(playerId, username string) {
	s.usernameCache[playerId] = username
}

func (s *Socket) uncacheUser(playerId string) {
	delete(s.usernameCache, playerId)
}

func (s *Socket) baseURL(websocket bool) string {
	if websocket {
		if s.ssl {
			return "wss://" + s.domain
		} else {
			return "ws://" + s.domain
		}
	} else {
		if s.ssl {
			return "https://" + s.domain
		} else {
			return "http://" + s.domain
		}
	}
}
