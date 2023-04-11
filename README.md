# Go-Client
![CodeGame Version](https://img.shields.io/badge/CodeGame-v0.8-orange)
![Go version](https://img.shields.io/github/go-mod/go-version/code-game-project/go-client)
[![GoDoc](https://pkg.go.dev/badge/github.com/code-game-project/go-client)](https://pkg.go.dev/github.com/code-game-project/go-client)

The Go client library for [CodeGame](https://code-game.org).

## Installation

```sh
go get github.com/code-game-project/go-client/cg
```

## Usage

```go
// Create a new game socket.
socket, err := cg.NewSocket("games.code-game.org/example")

// Create a new private and unprotected game on the server.
socket.CreateGame(false, false, nil)

// Join a game.
socket.Join(gameId, "username", "")

// Spectate a game.
socket.Spectate(gameId)

// Connect with an existing session.
socket.RestoreSession(username)

// Register an event listener for the `my_event` event.
socket.On(MyEventEvent, func(event cg.Event) {
	// TODO: do something with event
})

// Send a `hello_world` command.
socket.Send(HelloWorldCmd, HelloWorldCmdData{
	Message: "Hello, World!",
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
```

## License

MIT License

Copyright (c) 2022-2023 Julian Hofmann

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
