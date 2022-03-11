# Go-Client

## About

This is the Go client for [CodeGame](https://code-game-project.github.io/).

## Usage

```go
package main

import (
	"log"

	"github.com/code-game-project/go-client/cg"
)

func main() {
	// Open connection with CodeGame server.
	con, err := cg.Connect("ws://example-game.code-game.example.com/ws", "username")
	if err != nil {
		log.Fatalf("failed to connect to server: %s", err)
	}

	// Start listening for events.
	con.Listen()
}
```
