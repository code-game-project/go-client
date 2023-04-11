package cg

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

type Session struct {
	GameURL      string `json:"-"`
	Username     string `json:"-"`
	GameID       string `json:"game_id"`
	PlayerID     string `json:"player_id"`
	PlayerSecret string `json:"player_secret"`
}

var gamesPath = filepath.Join(xdg.DataHome, "codegame", "games")

func newSession(gameURL, username, gameID, playerID, playerSecret string) Session {
	return Session{
		GameURL:      gameURL,
		Username:     username,
		GameID:       gameID,
		PlayerID:     playerID,
		PlayerSecret: playerSecret,
	}
}

func loadSession(gameURL, username string) (Session, error) {
	data, err := os.ReadFile(filepath.Join(gamesPath, url.PathEscape(gameURL), username+".json"))
	if err != nil {
		return Session{}, err
	}

	var session Session
	err = json.Unmarshal(data, &session)

	session.GameURL = gameURL
	session.Username = username

	return session, err
}

func (s Session) save() error {
	if s.GameURL == "" {
		return errors.New("empty game url")
	}
	dir := filepath.Join(gamesPath, url.PathEscape(s.GameURL))
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, s.Username+".json"), data, 0o644)
}

func (s Session) remove() error {
	if s.GameURL == "" {
		return nil
	}
	dir := filepath.Join(gamesPath, url.PathEscape(s.GameURL))
	err := os.Remove(filepath.Join(dir, s.Username+".json"))
	if err != nil {
		return err
	}

	files, err := os.ReadDir(dir)
	if err == nil && len(files) == 0 {
		os.Remove(dir)
	}
	return nil
}
