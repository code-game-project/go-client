package cg

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/adrg/xdg"
)

type State struct {
	Name         string `json:"-"`
	GameId       string `json:"game_id"`
	PlayerId     string `json:"player_id"`
	PlayerSecret string `json:"player_secret"`
}

func RestoreState(name string) (State, error) {
	data, err := os.ReadFile(path.Join(xdg.DataHome, "CodeGame", name+".json"))
	if err != nil {
		return State{}, err
	}

	var state State
	err = json.Unmarshal(data, &state)

	state.Name = name

	return state, err
}

func (s State) Save() error {
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

func (s State) Remove() error {
	if s.Name == "" {
		return nil
	}
	return os.Remove(path.Join(xdg.DataHome, "CodeGame", s.Name+".json"))
}
