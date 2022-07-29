package cg

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type DebugSeverity string

const (
	DebugError   = "error"
	DebugWarning = "warning"
	DebugInfo    = "info"
	DebugTrace   = "trace"
)

type debugMessage struct {
	Severity DebugSeverity   `json:"severity"`
	Message  string          `json:"message"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// The data argument is empty if no data was included in the message.
type DebugMessageCallback func(severity DebugSeverity, message string, data string)

type DebugSocket struct {
	wsConn    *websocket.Conn
	callbacks map[CallbackId]DebugMessageCallback
	url       string
	tls       bool

	enableTrace   bool
	enableInfo    bool
	enableWarning bool
	enableError   bool
}

func NewDebugSocket(url string) *DebugSocket {
	url = trimURL(url)
	return &DebugSocket{
		callbacks:     make(map[CallbackId]DebugMessageCallback),
		url:           url,
		tls:           isTLS(url),
		enableTrace:   false,
		enableInfo:    true,
		enableWarning: true,
		enableError:   true,
	}
}

func (s *DebugSocket) URL() string {
	return s.url
}

// SetSeverities enables/disables specific message severities.
// SetSeverities panics if it is called after calling DebugServer, DebugGame or DebugPlayer.
// When SetSeverities is never called all severities except trace are enabled.
func (s *DebugSocket) SetSeverities(enableTrace, enableInfo, enableWarning, enableError bool) {
	if s.wsConn != nil {
		panic("cannot call SetSeverities after a connection has already been established")
	}
	s.enableTrace = enableTrace
	s.enableInfo = enableInfo
	s.enableWarning = enableWarning
	s.enableError = enableError
}

func (s *DebugSocket) OnMessage(callback DebugMessageCallback) CallbackId {
	id := CallbackId(uuid.New())
	s.callbacks[id] = callback
	return id
}

func (s *DebugSocket) RemoveCallback(id CallbackId) {
	delete(s.callbacks, id)
}

// DebugServer connects to the /api/debug endpoint on the server and listens for debug messages.
func (s *DebugSocket) DebugServer() error {
	wsConn, _, err := websocket.DefaultDialer.Dial(baseURL("ws", s.tls, "%s/api/debug?trace=%t&info=%t&warning=%t&error=%t", s.url, s.enableTrace, s.enableInfo, s.enableWarning, s.enableError), nil)
	if err != nil {
		return err
	}

	s.wsConn = wsConn

	return s.listen()
}

// DebugGame connects to the /api/games/{gameId}/debug endpoint on the server and listens for debug messages.
func (s *DebugSocket) DebugGame(gameId string) error {
	wsConn, _, err := websocket.DefaultDialer.Dial(baseURL("ws", s.tls, "%s/api/games/%s/debug?trace=%t&info=%t&warning=%t&error=%t", s.url, gameId, s.enableTrace, s.enableInfo, s.enableWarning, s.enableError), nil)
	if err != nil {
		return err
	}

	s.wsConn = wsConn

	return s.listen()
}

// DebugPlayer connects to the /api/games/{gameId}/players/{playerId}/debug endpoint on the server and listens for debug messages.
func (s *DebugSocket) DebugPlayer(gameId, playerId, playerSecret string) error {
	wsConn, _, err := websocket.DefaultDialer.Dial(baseURL("ws", s.tls, "%s/api/games/%s/players/%s/debug?player_secret=%s&trace=%t&info=%t&warning=%t&error=%t", s.url, gameId, playerId, playerSecret, s.enableTrace, s.enableInfo, s.enableWarning, s.enableError), nil)
	if err != nil {
		return err
	}

	s.wsConn = wsConn

	return s.listen()
}

func (s *DebugSocket) listen() error {
	for {
		msgType, msg, err := s.wsConn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseGoingAway) {
				return ErrClosed
			} else {
				return err
			}
		}
		if msgType != websocket.TextMessage {
			return ErrInvalidMessageType
		}

		var message debugMessage
		err = json.Unmarshal(msg, &message)
		if err != nil {
			return ErrDecodeFailed
		}

		dataStr := string(message.Data)
		for _, cb := range s.callbacks {
			cb(message.Severity, message.Message, dataStr)
		}
	}

}

// Close closes the underlying websocket connection.
func (s *DebugSocket) Close() error {
	s.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(5*time.Second))
	return s.wsConn.Close()
}
