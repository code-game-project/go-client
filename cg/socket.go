package cg

import (
	"encoding/json"
	"errors"
	"strconv"
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
	info           cgInfo
	url            string
	tls            bool
	session        Session
	wsConn         *websocket.Conn
	eventListeners map[EventName]map[CallbackId]EventCallback
	usernameCache  map[string]string

	running   bool
	eventChan chan Event
	err       error
}

// NewSocket creates a new Socket object ready to execute further actions.
// You can omit the protocol. NewSocket will determine the best protocol to use.
func NewSocket(url string) (*Socket, error) {
	url = trimURL(url)
	socket := &Socket{
		url:            url,
		tls:            isTLS(url),
		eventListeners: make(map[EventName]map[CallbackId]EventCallback),
		usernameCache:  make(map[string]string),
		eventChan:      make(chan Event, 10),
	}

	info, err := socket.fetchInfo()
	if err != nil {
		return nil, err
	}
	socket.info = info

	if !isVersionCompatible(info.CGVersion) {
		printWarning("CodeGame version mismatch. Server: v%s, client: v%s", info.CGVersion, CGVersion)
	}

	return socket, nil
}

// CreateGame creates a new game on the server and returns the id of the created game.
func (s *Socket) CreateGame(public bool) (string, error) {
	return s.createGame(public)
}

// Join creates a new player in the game and connects to it.
// Join panics if the socket is already connected to a game.
func (s *Socket) Join(gameId, username string) error {
	if s.session.GameURL != "" {
		panic("already connected to a game")
	}

	playerId, playerSecret, err := s.createPlayer(gameId, username)
	if err != nil {
		return err
	}

	return s.Connect(gameId, playerId, playerSecret)
}

// RestoreSession tries to restore the session and use it to reconnect to the game.
// RestoreSession panics if the socket is already connected to a game.
func (s *Socket) RestoreSession(username string) error {
	if s.session.GameURL != "" {
		panic("already connected to a game")
	}
	session, err := loadSession(s.url, username)
	if err != nil {
		return err
	}
	err = s.Connect(session.GameId, session.PlayerId, session.PlayerSecret)
	if err != nil {
		session.remove()
	}
	return err
}

// Connect connects to a game and player on the server.
// Connect panics if the socket is already connected to a game.
func (s *Socket) Connect(gameId, playerId, playerSecret string) error {
	err := s.connect(gameId, playerId, playerSecret)
	if err != nil {
		return err
	}

	s.session = newSession(s.url, "", gameId, playerId, playerSecret)

	s.startListenLoop()

	username, err := s.fetchUsername(gameId, playerId)
	if err != nil {
		return err
	}
	s.usernameCache[playerId] = username

	s.session.Username = username
	err = s.session.save()
	if err != nil {
		printError("Failed to save session: %s", err)
	}

	return nil
}

// Spectate joins the game as a spectator.
// Spectate panics if the socket is already connected to a game.
func (s *Socket) Spectate(gameId string) error {
	err := s.spectate(gameId)
	if err != nil {
		return err
	}
	s.session = newSession(s.url, "", gameId, "", "")
	s.startListenLoop()
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
func (s *Socket) On(event EventName, callback EventCallback) CallbackId {
	if s.eventListeners[event] == nil {
		s.eventListeners[event] = make(map[CallbackId]EventCallback)
	}

	id := CallbackId(uuid.New())

	s.eventListeners[event][id] = callback

	return id
}

// Once registers a callback that is triggered only the first time the event is received.
func (s *Socket) Once(event EventName, callback EventCallback) CallbackId {
	if s.eventListeners[event] == nil {
		s.eventListeners[event] = make(map[CallbackId]EventCallback)
	}

	id := CallbackId(uuid.New())

	s.eventListeners[event][id] = func(event Event) {
		callback(event)
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

// Send sends a new command to the server.
// Send panics if the socket is not connected to a player.
func (s *Socket) Send(name CommandName, data any) error {
	if s.wsConn == nil || s.session.PlayerId == "" {
		panic("not connected to a player")
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
func (s *Socket) Username(playerId string) string {
	if username, ok := s.usernameCache[playerId]; ok {
		return username
	}

	username, err := s.fetchUsername(s.session.GameId, playerId)
	if err == nil {
		s.usernameCache[playerId] = username
	}
	return username
}

// Session returns details of the current session.
func (s *Socket) Session() Session {
	return s.session
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
	if listeners != nil {
		for _, cb := range listeners {
			cb(event)
		}
	}
}

func isVersionCompatible(serverVersion string) bool {
	serverParts := strings.Split(serverVersion, ".")
	if len(serverParts) == 1 {
		serverParts = append(serverParts, "0")
	}

	clientParts := strings.Split(CGVersion, ".")
	if len(clientParts) == 1 {
		clientParts = append(clientParts, "0")
	}

	if serverParts[0] != clientParts[0] {
		return false
	}

	if clientParts[0] == "0" {
		return serverParts[1] == clientParts[1]
	}

	serverMinor, err := strconv.Atoi(serverParts[1])
	if err != nil {
		return false
	}

	clientMinor, err := strconv.Atoi(clientParts[1])
	if err != nil {
		return false
	}

	return clientMinor >= serverMinor
}
