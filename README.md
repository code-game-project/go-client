# Go-Client

## About

This is the Go client for [CodeGame](https://code-game-project.github.io/) v0.0.5.

## Installation

```sh
go get github.com/code-game-project/go-client/cg
```

## Usage

```go
package main

import (
	"log"

	// Import CodeGame client library.
	"github.com/code-game-project/go-client/cg"
)

func main() {
	// Open a websocket connection with CodeGame server.
	socket, err := cg.NewSocket("test", "ws://127.0.0.1:8080")
	if err != nil {
		log.Fatalf("failed to connect to server: %s", err)
	}

	// Register error event listener.
	socket.On(cg.EventError, func(origin string, target cg.EventTarget, event cg.Event) {
		var data cg.EventErrorData
		event.UnmarshalData(&data)
		log.Printf("server error: %s", data.Reason)
	})

	// Register a game_info event listener.
	socket.On(cg.EventGameInfo, func(origin string, target cg.EventTarget, event cg.Event) {
		var data cg.EventGameInfoData
		event.UnmarshalData(&data)
		fmt.Println(origin, target, event.Name, data)
	})

	// Register a game_info event listener, which is only triggered once.
	socket.OnOnce(cg.EventGameInfo, func(origin string, target cg.EventTarget, event cg.Event) {
		fmt.Println("GameInfoOnce")
	})

	// Try to restore the previous state (gameId, playerId, playerSecret).
	state, err := cg.RestoreState("test")
	if err == nil {
		// Connect to the game still stored in state.
		err = socket.Connect(state.GameId, state.PlayerId, state.PlayerSecret)
	}

	if err != nil {
		// Create a new game and store its id in 'gameId'.
		gameId, err := socket.Create()
		if err != nil {
			log.Fatalf("failed to create game: %s", err)
		}

		// Join the previously created game.
		state, err = socket.Join(gameId, "username")
		if err != nil {
			log.Fatalf("failed to join game: %s", err)
		}
	}

	// Save the current state (gameId, playerId, playerSecret).
	err = state.Save()
	if err != nil {
		log.Printf("failed to save state: %s", err)
	}

	// Start listening for events.
	err = socket.Listen()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}
```
