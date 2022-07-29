package cg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type cgInfo struct {
	Name          string `json:"name"`
	CGVersion     string `json:"cg_version"`
	DisplayName   string `json:"display_name"`
	Description   string `json:"description"`
	Version       string `json:"version"`
	RepositoryURL string `json:"repository_url"`
}

func (s *Socket) fetchInfo() (cgInfo, error) {
	resp, err := http.Get(baseURL("http", s.tls, "%s/api/info", s.url))
	if err != nil {
		return cgInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return cgInfo{}, fmt.Errorf("invalid response. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	var info cgInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return cgInfo{}, err
	}
	if info.Name == "" {
		return cgInfo{}, errors.New("empty `name` field")
	}
	if info.CGVersion == "" {
		return cgInfo{}, errors.New("empty `cg_version` field")
	}
	return info, err
}

func (s *Socket) createGame(public bool, config any) (string, error) {
	type request struct {
		Public bool `json:"public"`
		Config any  `json:"config,omitempty"`
	}
	data, err := json.Marshal(request{
		Public: public,
		Config: config,
	})
	if err != nil {
		return "", err
	}

	body := bytes.NewBuffer(data)
	resp, err := http.Post(baseURL("http", s.tls, "%s/api/games", s.url), "application/json", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("invalid response code: expected: %d, got: %d", http.StatusCreated, resp.StatusCode)
	}

	type response struct {
		GameId string `json:"game_id"`
	}
	var r response
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.GameId, err
}

func (s *Socket) createPlayer(gameId, username string) (string, string, error) {
	type request struct {
		Username string `json:"username"`
	}
	data, err := json.Marshal(request{
		Username: username,
	})
	if err != nil {
		return "", "", err
	}

	body := bytes.NewBuffer(data)
	resp, err := http.Post(baseURL("http", s.tls, "%s/api/games/%s/players", s.url, gameId), "application/json", body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("invalid response code: expected: %d, got: %d", http.StatusCreated, resp.StatusCode)
	}

	type response struct {
		PlayerId     string `json:"player_id"`
		PlayerSecret string `json:"player_secret"`
	}
	var r response
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.PlayerId, r.PlayerSecret, err
}

func (s *Socket) connect(gameId, playerId, playerSecret string) error {
	wsConn, _, err := websocket.DefaultDialer.Dial(baseURL("ws", s.tls, "%s/api/games/%s/connect?player_id=%s&player_secret=%s", s.url, gameId, playerId, playerSecret), nil)
	if err != nil {
		return err
	}
	s.wsConn = wsConn
	return nil
}

func (s *Socket) spectate(gameId string) error {
	wsConn, _, err := websocket.DefaultDialer.Dial(baseURL("ws", s.tls, "%s/api/games/%s/spectate", s.url, gameId), nil)
	if err != nil {
		return err
	}
	s.wsConn = wsConn
	return nil
}

func (s *Socket) fetchUsername(gameId, playerId string) (string, error) {
	resp, err := http.Get(baseURL("http", s.tls, "%s/api/games/%s/players/%s", s.url, gameId, playerId))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid response. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	type response struct {
		Username string `json:"username"`
	}
	var r response
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.Username, err
}
