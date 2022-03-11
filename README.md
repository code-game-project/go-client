# Go-Client

## About

This is the Go client for [CodeGame](https://code-game-project.github.io/).

## Usage

```go
package main

import (
	"log"

	// Import CodeGame client library.
	"github.com/code-game-project/go-client/cg"
)

func main() {
	// Open connection with CodeGame server.
	con, err := cg.Connect("ws://127.0.0.1:8081/ws", "username")
	if err != nil {
		log.Fatalf("failed to connect to server: %s", err)
	}

	// Register error event listener.
	con.On(cg.ErrorEvent, func(origin string, target cg.EventTarget, event cg.Event) {
		// Decode event data into the data struct.
		var data cg.ErrorEventData
		event.UnmarshalData(&data)

		// Log error.
		if origin == cg.EventOriginSelf {
			// Error originated in this client.
			log.Printf("error: %s", data.Reason)
		} else {
			// Error originated in server.
			log.Printf("server error: %s", data.Reason)
		}
	})

	// Start listening for events.
	con.Listen()
}
```
