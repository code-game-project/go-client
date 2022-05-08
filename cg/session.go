package cg

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/adrg/xdg"
)

type Session struct {
	Name         string `json:"-"`
	GameId       string `json:"game_id"`
	PlayerId     string `json:"player_id"`
	PlayerSecret string `json:"player_secret"`
}

func RestoreSession(name string) (Session, error) {
	data, err := os.ReadFile(path.Join(xdg.DataHome, "codegame", name+".json"))
	if err != nil {
		return Session{}, err
	}

	var session Session
	err = json.Unmarshal(data, &session)

	session.Name = name

	return session, err
}

func (s Session) Save() error {
	if s.Name == "" {
		return errors.New("empty name")
	}
	dir := path.Join(xdg.DataHome, "CodeGame")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(dir, s.Name+".json"), data, 0644)
}

func (s Session) Remove() error {
	if s.Name == "" {
		return nil
	}
	return os.Remove(path.Join(xdg.DataHome, "CodeGame", s.Name+".json"))
}
