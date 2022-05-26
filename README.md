# Go-Client
![CG Protocol Version](https://img.shields.io/badge/Protocol-v0.6-orange)
![CG Client Version](https://img.shields.io/badge/Client-v0.3-yellow)
![Go version](https://img.shields.io/github/go-mod/go-version/code-game-project/go-client)

This is the Go client library for [CodeGame](https://github.com/code-game-project).

## Installation

```sh
go get github.com/code-game-project/go-client/cg
```

## [Getting Started](./docs/GETTING_STARTED.md)

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
	socket, err := cg.NewSocket("localhost:8080")
	if err != nil {
		log.Fatalf("failed to connect to server: %s", err)
	}

	// Try to connect with a previous session.
	err = socket.RestoreSession("username")
	if err != nil {
		// Create a new private game and store its id in 'gameId'.
		gameId, err := socket.Create(false)
		if err != nil {
			log.Fatalf("failed to create game: %s", err)
		}

		// Join the previously created game.
		err = socket.Join(gameId, "username")
		if err != nil {
			log.Fatalf("failed to join game: %s", err)
		}
	}

	// Register error event listener.
	socket.On(cg.EventError, func(origin string, event cg.Event) {
		var data cg.EventErrorData
		event.UnmarshalData(&data)
		log.Printf("server error: %s", data.Message)
	})

	// Start listening for events. Blocks until the connection is closed.
	err = socket.RunEventLoop()
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	// ======= ALTERNATIVELY =======

	// manual event loop
	for {
		// NextEvent returns the next event in the queue or ok = false if there is none.
		// Registered event listeners will be triggered.
		// event -> The polled event. Only valid if ok == true.
		// ok -> Whether there was an event in the queue
		// err -> cg.ErrClosed if closed.
		event, ok, err := socket.NextEvent()
		if err != nil {
			if err != cg.ErrClosed {
				log.Fatalf("error: %s", err)
			}
			break
		}
		if ok {
			// do something with event
		}
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
