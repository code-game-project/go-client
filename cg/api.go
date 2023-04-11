package cg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

type CGInfo struct {
	Name          string `json:"name"`
	CGVersion     string `json:"cg_version"`
	DisplayName   string `json:"display_name"`
	Description   string `json:"description"`
	Version       string `json:"version"`
	RepositoryURL string `json:"repository_url"`
}

func (s *Socket) fetchInfo() (CGInfo, error) {
	resp, err := http.Get(baseURL("http", s.tls, "%s/api/info", s.url))
	if err != nil {
		return CGInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err == nil && len(data) > 0 {
			return CGInfo{}, fmt.Errorf("Failed to fetch game info: %s", string(data))
		}
		return CGInfo{}, fmt.Errorf("invalid response. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	var info CGInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return CGInfo{}, err
	}
	if info.Name == "" {
		return CGInfo{}, errors.New("empty `name` field")
	}
	if info.CGVersion == "" {
		return CGInfo{}, errors.New("empty `cg_version` field")
	}
	return info, err
}

func (s *Socket) createGame(public, protected bool, config any) (gameID string, joinSecret string, err error) {
	type request struct {
		Public    bool `json:"public"`
		Protected bool `json:"protected"`
		Config    any  `json:"config,omitempty"`
	}
	data, err := json.Marshal(request{
		Public:    public,
		Protected: protected,
		Config:    config,
	})
	if err != nil {
		return "", "", err
	}

	body := bytes.NewBuffer(data)
	resp, err := http.Post(baseURL("http", s.tls, "%s/api/games", s.url), "application/json", body)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("invalid response code: expected: %d, got: %d", http.StatusCreated, resp.StatusCode)
	}

	type response struct {
		GameID     string `json:"game_id"`
		JoinSecret string `json:"join_secret"`
	}
	var r response
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.GameID, r.JoinSecret, err
}

func (s *Socket) createPlayer(gameID, username, joinSecret string) (string, string, error) {
	type request struct {
		Username   string `json:"username"`
		JoinSecret string `json:"join_secret,omitempty"`
	}
	data, err := json.Marshal(request{
		Username:   username,
		JoinSecret: joinSecret,
	})
	if err != nil {
		return "", "", err
	}

	body := bytes.NewBuffer(data)
	resp, err := http.Post(baseURL("http", s.tls, "%s/api/games/%s/players", s.url, gameID), "application/json", body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusCreated {
		data, err := io.ReadAll(resp.Body)
		if err == nil && len(data) > 0 {
			return "", "", errors.New(string(data))
		}
		return "", "", fmt.Errorf("invalid response code: expected: %d, got: %d", http.StatusCreated, resp.StatusCode)
	}

	type response struct {
		PlayerID     string `json:"player_id"`
		PlayerSecret string `json:"player_secret"`
	}
	var r response
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.PlayerID, r.PlayerSecret, err
}

func (s *Socket) connect(gameID, playerID, playerSecret string) error {
	wsConn, _, err := websocket.DefaultDialer.Dial(baseURL("ws", s.tls, "%s/api/games/%s/players/%s/connect?player_secret=%s", s.url, gameID, playerID, playerSecret), nil)
	if err != nil {
		return err
	}
	s.wsConn = wsConn
	return nil
}

func (s *Socket) spectate(gameID string) error {
	wsConn, _, err := websocket.DefaultDialer.Dial(baseURL("ws", s.tls, "%s/api/games/%s/spectate", s.url, gameID), nil)
	if err != nil {
		return err
	}
	s.wsConn = wsConn
	return nil
}

func (s *Socket) fetchUsername(gameID, playerID string) (string, error) {
	resp, err := http.Get(baseURL("http", s.tls, "%s/api/games/%s/players/%s", s.url, gameID, playerID))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err == nil && len(data) > 0 {
			return "", fmt.Errorf("Failed to fetch username of %s: %s", playerID, string(data))
		}
		return "", fmt.Errorf("invalid response. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	type response struct {
		Username string `json:"username"`
	}
	var r response
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.Username, err
}

func (s *Socket) fetchPlayers(gameID string) (map[string]string, error) {
	resp, err := http.Get(baseURL("http", s.tls, "%s/api/games/%s/players", s.url, gameID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err == nil && len(data) > 0 {
			return nil, fmt.Errorf("Failed to fetch players: %s", string(data))
		}
		return nil, fmt.Errorf("invalid response. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
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
	resp, err := http.Get(baseURL("http", socket.tls, "%s/api/games/%s", socket.url, gameID))
	if err != nil {
		return config, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err == nil && len(data) > 0 {
			return config, fmt.Errorf("Failed to fetch game config: %s", string(data))
		}
		return config, fmt.Errorf("invalid response. expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	var r configReponse[T]
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.Config, err
}
