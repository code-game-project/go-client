# Go-Client
![CodeGame Version](https://img.shields.io/badge/CodeGame-v0.1-orange)
![Go version](https://img.shields.io/github/go-mod/go-version/code-game-project/go-client)

This is the Go client library for [CodeGame](https://github.com/code-game-project).

## Installation

```sh
go get github.com/code-game-project/go-client/cg
```

## Usage

```go
package main

import (
	"fmt"
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
	socket.On(cg.EventError, func(origin string, event cg.Event) {
		var data cg.EventErrorData
		event.UnmarshalData(&data)
		log.Printf("server error: %s", data.Reason)
	})

	// Register a game_info event listener.
	socket.On(cg.EventGameInfo, func(origin string, event cg.Event) {
		var data cg.EventGameInfoData
		event.UnmarshalData(&data)
		fmt.Println(origin, event.Name, data)
	})

	// Register a game_info event listener, which is only triggered once.
	socket.OnOnce(cg.EventGameInfo, func(origin string, event cg.Event) {
		fmt.Println("GameInfoOnce")
	})

	// Try to restore the previous state (gameId, playerId, playerSecret).
	state, err := cg.RestoreState("test")
	if err == nil {
		// Connect to the game still stored in state.
		err = socket.Connect(state.GameId, state.PlayerId, state.PlayerSecret)
	}

	if err != nil {
		// Create a new private game and store its id in 'gameId'.
		gameId, err := socket.Create(false)
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

## License

MIT License

Copyright (c) 2022 CodeGame Contributors (https://github.com/orgs/code-game-project/people)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
