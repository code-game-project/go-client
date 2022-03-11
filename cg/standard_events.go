package cg

// The create_game event is used to create a new game.
const CreateGameEvent = "create_game"

type CreateGameEventData struct {
	Username string `json:"username"`
}

// The created_game event is the server’s response to the create_game event. It is only sent to the client that created the game.
const CreatedGameEvent = "created_game"

type CreatedGameEventData struct {
	GameId string `json:"game_id"`
}

// The join_game event is used to join an existing game by id.
const JoinGameEvent = "join_game"

type JoinGameEventData struct {
	GameId   string `json:"game_id"`
	Username string `json:"username"`
}

// The joined_game event is sent to everyone in the game when someone joins it.
const JoinedGameEvent = "joined_game"

type JoinedGameEventData struct {
	Username string `json:"username"`
}

// The leave_game event is used to leave a game which is the preferred way to exit a game in comparisson to just disconnecting and never reconnecting.
// It is not required to send this event due to how hard it is to detect if the user has disconnected for good or is just re-writing their program.
const LeaveGameEvent = "leave_game"

type LeaveGameEventData struct{}

// The left_game event is sent to everyone in the game when someone leaves it.
const LeftGameEvent = "left_game"

type LeftGameEventData struct{}

// The disconnected event is sent to everyone in the game when someone disconnects from the server.
const DisconnectedEvent = "disconnected"

type DisconnectedEventData struct{}

// The reconnected event is sent to everyone in the game when someone reconnects to the server.
const ReconnectedEvent = "reconnected"

type ReconnectedEventData struct{}

// The game_info event is sent to every player that joins a game with 1 or more players.
const GameInfoEvent = "game_info"

type GameInfoEventData struct {
	Players map[string]string `json:"players"`
}

// The error event is sent to the client of the server socket instance where the error occurred. If the error affects multiple players it should be sent to all of the affected players.
const ErrorEvent = "error"

type ErrorEventData struct {
	Reason string `json:"reason"`
}
