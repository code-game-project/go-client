package cg

import (
	"encoding/json"

	"github.com/google/uuid"
)

// A wrapper struct around and event and its origin.
type EventWrapper struct {
	Origin string `json:"origin"`
	Event  Event  `json:"event"`
}

type CallbackId uuid.UUID
type OnEventCallback func(origin string, event Event)

type Event struct {
	Name EventName       `json:"name"`
	Data json.RawMessage `json:"data"`
}

// UnmarshalData decodes the event data into the struct pointed to by targetObjPtr.
func (e *Event) UnmarshalData(targetObjPtr any) error {
	return json.Unmarshal(e.Data, targetObjPtr)
}

// marshalData encodes obj into the Data field of the event.
func (e *Event) marshalData(obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	e.Data = data
	return nil
}
