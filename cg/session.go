package cg

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

type Session struct {
	Name     string `json:"-"`
	Username string `json:"-"`
	GameId   string `json:"game_id"`
	PlayerId string `json:"player_id"`
	Secret   string `json:"secret"`
	Path     string `json:"-"`
}

var gamesPath = filepath.Join(xdg.DataHome, "codegame", "games")

func newSession(name, username, gameId, playerId, secret string) Session {
	return Session{
		Name:     name,
		Username: username,
		GameId:   gameId,
		PlayerId: playerId,
		Secret:   secret,
	}
}

func loadSession(name, username string) (Session, error) {
	data, err := os.ReadFile(filepath.Join(gamesPath, name, username+".json"))
	if err != nil {
		return Session{}, err
	}

	var session Session
	err = json.Unmarshal(data, &session)

	session.Name = name
	session.Username = username

	return session, err
}

func (s Session) save() error {
	if s.Name == "" {
		return errors.New("empty name")
	}
	dir := filepath.Join(gamesPath, s.Name)
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
	if s.Name == "" {
		return nil
	}
	return os.Remove(filepath.Join(gamesPath, s.Name, s.Username+".json"))
}
