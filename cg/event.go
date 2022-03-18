package cg

import "encoding/json"

type EventTargetType string

const (
	EventTargetTypeGame   EventTargetType = "game"
	EventTargetTypeSocket EventTargetType = "socket"
	EventTargetTypeSelf   EventTargetType = "self"
)

type EventTarget struct {
	Type EventTargetType `json:"type"`
	ID   string          `json:"id"`
}

const (
	EventOriginServer = "server"
	EventOriginSelf   = "self"
)

type eventWrapper struct {
	Target EventTarget `json:"target"`
	Origin string      `json:"origin"`
	Event  Event       `json:"event"`
}

type OnEventCallback func(origin string, target EventTarget, event Event)

type Event struct {
	Name EventName       `json:"name"`
	Data json.RawMessage `json:"data"`
}

// UnmarshalData decodes the event data into the struct pointed to by targetObjPtr.
func (e *Event) UnmarshalData(targetObjPtr interface{}) error {
	return json.Unmarshal(e.Data, targetObjPtr)
}

// marshalData encodes obj into the Data field of the event.
func (e *Event) marshalData(obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	e.Data = data
	return nil
}
