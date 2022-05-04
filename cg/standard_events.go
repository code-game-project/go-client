package cg

type EventName string

// The create_game event is used to create a new game.
const EventCreateGame EventName = "create_game"

type EventCreateGameData struct{}

// The created_game event is the server’s response to the create_game event. It is only sent to the client that created the game.
const EventCreatedGame EventName = "created_game"

type EventCreatedGameData struct {
	GameId string `json:"game_id"`
}

// The join_game event is used to join an existing game by id.
const EventJoinGame EventName = "join_game"

type EventJoinGameData struct {
	GameId   string `json:"game_id"`
	Username string `json:"username"`
}

// The joined_game event is sent to everyone in the game when someone joins it.
const EventJoinedGame EventName = "joined_game"

type EventJoinedGameData struct {
	Username string `json:"username"`
}

// The player_secret event is used to send a secret to the player that just joined so that they can reconnect and add other clients..
const EventPlayerSecret EventName = "player_secret"

type EventPlayerSecretData struct {
	Secret string `json:"secret"`
}

// The leave_game event is used to leave a game which is the preferred way to exit a game in comparisson to just disconnecting and never reconnecting.
// It is not required to send this event due to how hard it is to detect if the user has disconnected for good or is just re-writing their program.
const EventLeaveGame EventName = "leave_game"

type EventLeaveGameData struct{}

// The left_game event is sent to everyone in the game when someone leaves it.
const EventLeftGame EventName = "left_game"

type EventLeftGameData struct{}

// The disconnected event is sent to everyone in the game when someone disconnects from the server.
const EventDisconnected EventName = "disconnected"

type EventDisconnectedData struct{}

// The connected event is used to associate a client with an existing player. This event is used after making changes to ones program and reconnecting to the game or when adding another client like a viewer in the webbrowser.
const EventConnect EventName = "connect"

type EventConnectData struct {
	GameId   string `json:"game_id"`
	PlayerId string `json:"player_id"`
	Secret   string `json:"secret"`
}

// The connected event is sent to everyone in the game when a player connects a client to the server.
const EventConnected EventName = "connected"

type EventConnectedData struct{}

// The game_info event is sent to every player that joins a game with 1 or more players.
const EventGameInfo EventName = "game_info"

type EventGameInfoData struct {
	Players map[string]string `json:"players"`
}

// The error event is sent to the client of the server socket instance where the error occurred. If the error affects multiple players it should be sent to all of the affected players.
const EventError EventName = "error"

type EventErrorData struct {
	Reason string `json:"reason"`
}

// Returns true if eventName is a standard event.
func IsStandardEvent(eventName EventName) bool {
	return eventName == EventCreateGame || eventName == EventCreatedGame || eventName == EventDisconnected ||
		eventName == EventError || eventName == EventGameInfo || eventName == EventJoinGame || eventName == EventJoinedGame ||
		eventName == EventLeaveGame || eventName == EventLeftGame || eventName == EventConnect || eventName == EventConnected ||
		eventName == EventPlayerSecret
}
