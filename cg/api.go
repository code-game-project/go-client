package cg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

func (s *Socket) connect(gameID, playerID, playerSecret string) error {
	wsConn, _, err := websocket.DefaultDialer.Dial(baseURL("ws", s.tls, "%s/api/games/%s/players/%s/connect?player_secret=%s", s.gameURL, gameID, playerID, playerSecret), nil)
	if err != nil {
		return err
	}
	s.wsConn = wsConn
	return nil
}

func (s *Socket) spectate(gameID string) error {
	wsConn, _, err := websocket.DefaultDialer.Dial(baseURL("ws", s.tls, "%s/api/games/%s/spectate", s.gameURL, gameID), nil)
	if err != nil {
		return err
	}
	s.wsConn = wsConn
	return nil
}

func (s *Socket) fetchUsername(gameID, playerID string) (string, error) {
	resp, err := http.Get(baseURL("http", s.tls, "%s/api/games/%s/players/%s", s.gameURL, gameID, playerID))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var data []byte
		data, err = io.ReadAll(resp.Body)
		if err == nil && len(data) > 0 {
			return "", fmt.Errorf("failed to fetch username of %s: %s", playerID, string(data))
		}
		return "", fmt.Errorf("invalid response; expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	type response struct {
		Username string `json:"username"`
	}
	var r response
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.Username, err
}

func (s *Socket) fetchPlayers(gameID string) (map[string]string, error) {
	resp, err := http.Get(baseURL("http", s.tls, "%s/api/games/%s/players", s.gameURL, gameID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var data []byte
		data, err = io.ReadAll(resp.Body)
		if err == nil && len(data) > 0 {
			return nil, fmt.Errorf("failed to fetch players: %s", string(data))
		}
		return nil, fmt.Errorf("invalid response; expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	var r map[string]string
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r, err
}

type configReponse[T any] struct {
	Config T `json:"config"`
}

// FetchGameConfig fetches the game config from the server.
func FetchGameConfig[T any](socket *Socket, gameID string) (T, error) {
	var config T
	resp, err := http.Get(baseURL("http", socket.tls, "%s/api/games/%s", socket.gameURL, gameID))
	if err != nil {
		return config, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var data []byte
		data, err = io.ReadAll(resp.Body)
		if err == nil && len(data) > 0 {
			return config, fmt.Errorf("failed to fetch game config: %s", string(data))
		}
		return config, fmt.Errorf("invalid response; expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	var r configReponse[T]
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.Config, err
}
