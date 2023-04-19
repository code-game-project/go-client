package cg

import (
	"encoding/json"
)

type (
	CallbackID    int64
	EventCallback func(event Event)
)

type EventName string

type Event struct {
	Name EventName       `json:"name"`
	Data json.RawMessage `json:"data"`
}

type CommandName string

type Command struct {
	Name CommandName     `json:"name"`
	Data json.RawMessage `json:"data"`
}

// UnmarshalData decodes the event data into the struct pointed to by targetObjPtr.
func (e *Event) UnmarshalData(targetObjPtr any) error {
	return json.Unmarshal(e.Data, targetObjPtr)
}

// marshalData encodes obj into the Data field of the command.
func (c *Command) marshalData(obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	c.Data = data
	return nil
}
