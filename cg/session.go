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
	GameId       string `json:"game_id"`
	PlayerId     string `json:"player_id"`
	PlayerSecret string `json:"player_secret"`
	Path         string `json:"-"`
}

var gamesPath = filepath.Join(xdg.DataHome, "codegame", "games")

func newSession(gameURL, username, gameId, playerId, playerSecret string) Session {
	return Session{
		GameURL:      gameURL,
		Username:     username,
		GameId:       gameId,
		PlayerId:     playerId,
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
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, s.Username+".json"), data, 0644)
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

	dirs, err := os.ReadDir(dir)
	if err == nil && len(dirs) == 0 {
		os.Remove(dir)
	}
	return nil
}
