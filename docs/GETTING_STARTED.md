# Getting Started

This guide will walk you through creating your own client for the [tic-tac-toe-simple](https://github.com/code-game-project/tic-tac-toe-simple)
game at `codegame.julianh.de/tic-tac-toe-simple`.
It is recommended that you first read the general CodeGame [getting started guide](https://github.com/code-game-project/.github/blob/main/getting-started.md).

## Scope

This guide will teach you the most important functions of this client library.
It does not focus on creating a beautiful and pleasant to use application. This task is up to you.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Setting up the project](#setting-up-the-project)
  - [Go module setup](#go-module-setup)
  - [Installing go-client](#installing-go-client)
  - [Generating the tic-tac-toe-simple event definitions](#generating-the-tic-tac-toe-simple-event-definitions)
- [Joining a game](#joining-a-game)
  - [Connecting to the game server](#connecting-to-the-game-server)
  - [Creating a game](#creating-a-game)
  - [Joining the game](#joining-the-game)
  - [Restore a session](#restoring-a-session)
  - [Putting it all together](#putting-it-all-together)
- [Listening for events](#listening-for-events)
- [Starting the event loop](#starting-the-event-loop)
- [Implementing the tic-tac-toe-simple client](#implementing-the-tic-tac-toe-simple-client)
  - [Handling events](#handling-events)
  - [The start event](#the-start-event)
  - [The invalid_action event](#the-invalid_action-event)
  - [The board event](#the-board-event)
  - [The turn event](#the-turn-event)
  - [The mark event](#the-mark-event)
  - [The game_over event](#the-game_over-event)
- [What next?](#what-next)

## Prerequisites

In order to follow this guide you will have to have the following software installed:

- [Go](https://go.dev) 1.18+

## Setting up the project

### Go module setup

The next step is to create a new Go project for our game client using the following commands:

```sh
mkdir tic-tac-toe-simple-client
cd tic-tac-toe-simple-client
go mod init tic-tac-toe-simple-client
```

After creating the directory create a simple `main.go` file:

```go
package main

func main() {

}
```

### Installing go-client

Now we need to install the [CodeGame client library for Go](https://github.com/code-game-project/go-client):

```sh
go get github.com/code-game-project/go-client/cg
```

The library can be imported into your code with an `import` statement:

```go
package main

// Import go-client.
import "github.com/code-game-project/go-client/cg"

func main() {}
```

### Generating the tic-tac-toe-simple event definitions

Tic-tac-toe-simple adds several new events to CodeGame.
In order for you to use these you will need to generate an `events.go` file with the [cg-gen-events](https://github.com/code-game-project/cg-gen-events) utility containing all of the event definitions you will need:

```sh
# cg-gen-events will automatically fetch the CGE file from the game server.
cg-gen-events https://codegame.julianh.de/tic-tac-toe-simple
```

You will probably need to change the package name in `events.go` to

```go
package main
```

## Joining a game

### Connecting to the game server

The first step in every CodeGame client is to open a connection with the game server.
With this client library it's as simple as calling the `cg.NewSocket` function with the URL of the game server:

```go
// You can omit the protocol. The client library will determine the best protocol to use.
socket, err := cg.NewSocket("codegame.julianh.de/tic-tac-toe-simple")
```

### Creating a game

Before we can join a game we need to create one.
This can be done with the `socket.Create` method:

```go
// The boolean value specifies whether the game should be public.
gameId, err := socket.Create(false)
```

### Joining the game

Once you have a game ID either provided by the user or by calling `socket.Create` you can use `socket.Join` to join a game.

```go
err := socket.Join(gameId, username)
```

### Restoring a session

If successful `socket.Join` will store the session on disk.
To restore the session after the application has been restarted use the `socket.RestoreSession` method:

```go
err = socket.RestoreSession(username)
```

### Putting it all together

These functions can be combined to let the user create or join a game, whilst reusing existing sessions.
This is achieved by letting the user provide a username and an optional game ID.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/code-game-project/go-client/cg"
)

func main() {
	// Print usage if the user has not supplied a username.
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s <username> <gameId?>\n", os.Args[0])
		os.Exit(1)
	}

	// Connect to the game server. This does not yet join any game.
	socket, err := cg.NewSocket("codegame.julianh.de/tic-tac-toe-simple")
	if err != nil {
		log.Fatal(err)
	}

	// Try to restore a previous session. If it fails run the code inside of the if branch.
	if err = socket.RestoreSession(os.Args[1]); err != nil {
		var gameId string
		// If the user has supplied a game ID, use it. Otherwise create a new game on the server.
		if len(os.Args) == 3 {
			gameId = os.Args[2]
		} else {
			gameId, err = socket.Create(false)
			if err != nil {
				log.Fatalf("failed to create game: %s", err)
			}
			fmt.Println("Game ID:", gameId)
		}

		// Join the game with the username provided as a command line argument.
		err = socket.Join(gameId, os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
	}
}
```

Now you can create a new game (or reconnect to an existing session):

```sh
go run . <username>
```

or join an existing game:

```sh
go run . <username> <game_id>
```

## Listening for events

You can register an event listener with the `socket.On` or the `socket.OnOnce` method.

For example to register an error event handler use the following snippet:

```go
// if err = socket.RestoreSession(os.Args[1]); err != nil {...}

// Register a callback which will be call every time the `cg_error` event is received.
socket.On(cg.EventError, func(origin string, event cg.Event) {
	// Inside of the event handler you receive the origin of the event (either 'server' or the ID of a player) and the event itself.

	// The event data is not yet usable. It first needs to be deserialized with these two lines of code:
	var data cg.EventErrorData
	event.UnmarshalData(&data)

	// Finally print the error message.
	fmt.Println("error:", data.Message)
})
```

The error event handler is usually registered after successfully joining/connecting to a game
because the previously mentioned methods return any error they might receive.

## Starting the event loop

At this point you will not receive any events because the client is not actually listening for them.
That is what an event loop is for: listening for new events and calling all event listeners.

There are two ways of writing an event loop:

1. Using `socket.RunEventLoop`
2. Writing your own loop calling `socket.NextEvent` repeatedly.

In this guide we will be using the first option because it is simpler.
The second option is most often needed when you are using a framework which provides its own loop like Unity, Raylib or similar.


To start the event loop in our tic-tac-toe-simple client add the following line of code after registering the error event handler.


```go
// socket.On(cg.EventError, func(origin string, event cg.Event) {...})

// socket.RunEventLoop will blook until the connection is closed.
err = socket.RunEventLoop()
// TODO handle err
```

## Implementing the tic-tac-toe-simple client

At this point we have a usable structure which allows us to easily implement the game.

### The client struct

Because we will need the `socket` variable throughout the application it is advised to implement all of the game logic as methods of a `client` struct which stores the socket and other useful state.

Let's declare and use the `client` struct:

```go
// package ...
// import (...)

// The client struct stores the CodeGame socket and the sign of the current player ('x' or 'o')
type client struct {
	socket *cg.Socket
	// Sign is defined in `events.go`
	sign   Sign
}

// Instead of calling `socket.RunEventLoop` directly in the main function we will call it in `client.run` after registering all needed event listeners.
func (c *client) run() error {
	// TODO Register event listeners.

	return c.socket.RunEventLoop()
}

func main() {
	// ...

	// Replace `err = socket.RunEventLoop` with:
	client := &client{
		socket: socket,
	}
	err = client.run()
	// TODO handle err
}
```

### Handling events

#### The `start` event

Once a match is found the server sends the `start` event which includes the player IDs mapped to their signs.
When we receive the event we deserialize the event data, store the sign in the `client` struct and print it out.

```go
func (c *client) run() error {
	c.socket.On(EventStart, func(origin string, event cg.Event) {
		// Deserialize the event data.
		var data EventStartData
		event.UnmarshalData(&data)

		// `socket.Session` returns a struct with useful information like the current game ID or the player ID.
		// In this case we need to player ID to receive the sign of our player.
		c.sign = data.Signs[c.socket.Session().PlayerId]

		// Print the sign.
		fmt.Println("Found a match! Your sign is:", c.sign)
	})

	// return c.socket.RunEventLoop()
}
```

#### The `invalid_action` event

Another useful but not required event is the `invalid_action` event which is sent when we do something wrong like trying to mark an already occupied field.
In this case we want to print the error message.

```go
func (c *client) run() error {
	// c.socket.On(EventStart, func(origin string, event cg.Event) {...})

	c.socket.On(EventInvalidAction, func(origin string, event cg.Event) {
		var data EventInvalidActionData
		event.UnmarshalData(&data)
		fmt.Println(data.Message)
	})

	// return c.socket.RunEventLoop()
}
```

#### The `board` event

The `board` event tells us the current state of the board.
Every time we receive this event we want to print the board to the console.

```go
func (c *client) run() error {
	// c.socket.On(EventInvalidAction, func(origin string, event cg.Event) {...})

	c.socket.On(EventBoard, func(origin string, event cg.Event) {
		var data EventBoardData
		event.UnmarshalData(&data)
		c.printBoard(data.Board)
	})

	// return c.socket.RunEventLoop()
}

func (c *client) printBoard(board [][]Field) {
	// Print a separator.
	fmt.Println(strings.Repeat("=", 50))

	// Loop through all rows.
	for i := range board {
		// Loop through all columns.
		for j := range board[i] {
			// Print a '/' symbol if the field is not occupied.
			if board[i][j].Sign == SignNone {
				fmt.Print("/")
			} else {
				// Otherwise print the sign on it.
				fmt.Print(board[i][j].Sign)
			}
		}
		// Start a new row.
		fmt.Print("\n")
	}
}
```

At this point you should be able to run an instance of the client with: `go run . <username>`, join the same game with a second client with `go run . <other_username> <game_id>` and see the game board in both consoles.

#### The `turn` event

What good is a board if we can't use it? Well, the `turn` event notifies us which player's turn it currently is so they can mark a field.

Once we receive a `turn` event we need to check whether it's our turn and let the player input a field if it is.

```go
func (c *client) run() error {
	// c.socket.On(EventBoard, func(origin string, event cg.Event) {...})

	c.socket.On(EventTurn, func(origin string, event cg.Event) {
		var data EventTurnData
		event.UnmarshalData(&data)

		if data.Sign == c.sign {
			// It's our turn.
			fmt.Println(strings.Repeat("=", 50))
			c.mark()
		} else {
			// It's not our turn.
			fmt.Println("Waiting for opponent…")
		}
	})

	// return c.socket.RunEventLoop()
}

func (c *client) mark() {
	// TODO
}
```

#### The `mark` event

There is only one event which we send to the server: the `mark` event. It allows us to mark an empty field with our sign provided it's our turn.

We already know when it's our turn and call the `mark` method so let's let the user input a field and send it to the server.

```go
func (c *client) mark() {

	// Ask the user to input a field (e.g. 1,1 for the top left field)
	fmt.Print("Where do you want to place your sign? (row,column) ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	location := scanner.Text()

	// TODO input validation

	// Split row and column.
	coords := strings.Split(location, ",")

	// Convert row and column to an integer.
	row, _ := strconv.Atoi(coords[0])
	column, _ := strconv.Atoi(coords[1])

	// Send the `mark` event with the row and column to the server.
	c.socket.Send(EventMark, EventMarkData{
		Row:    row - 1,
		Column: column - 1,
	})
}
```

We also probably want to let the user enter again if they made a mistake.
Because of that we call the `mark` method at the bottom of the `invalid_action` event callback:

```go
func (c *client) run() error {
	// c.socket.On(EventStart, func(origin string, event cg.Event) {...})

	c.socket.On(EventInvalidAction, func(origin string, event cg.Event) {
		var data EventInvalidActionData
		event.UnmarshalData(&data)
		fmt.Println(data.Message)
		c.mark() // <-------
	})

	// c.socket.On(EventBoard, func(origin string, event cg.Event) {...})
}
```

#### The `game_over` event

There is only one event left to go. The `game_over` event is sent once either all fields have been marked or a player won.

Apart from the type of ending and the winning sign the `game_over` event also returns the fields which form the winning row.
For simplicity we will only print the outcome.

```go
	// c.socket.On(EventTurn, func(origin string, event cg.Event) {...})

	c.socket.On(EventGameOver, func(origin string, event cg.Event) {
		fmt.Println(strings.Repeat("=", 50))
		var data EventGameOverData
		event.UnmarshalData(&data)

		// The boolean `tie` is true if it's a tie.
		if data.Tie {
			fmt.Println("Tie!")
		} else if data.WinnerSign == c.sign {
			// The current player wins if the winner sign matches the player sign.
			fmt.Println("You win!")
		} else {
			fmt.Println("You lose!")
		}
	})

	// return c.socket.RunEventLoop()
```

## What next?

With only about 170 lines of code our tic-tac-toe-simple client is finished. But where to go from here?

I recommend reading the following specifications to build a stronger understanding of CodeGame:

- [CodeGame Protocol Specification](https://github.com/code-game-project/docs/blob/main/protocol-specification.md) (Definitely useful)
- [CodeGame Events Language Specification](https://github.com/code-game-project/docs/blob/main/code-game-events-language-specification.md) (Useful if you want to write a game or understand existing games better)
- [CodeGame Game Server Specification](https://github.com/code-game-project/docs/blob/main/game-server-specification.md) (Mostly useful if you plan to write a client which displays more information like a list of public games from a server)
- [CodeGame Client Library Specification](https://github.com/code-game-project/docs/blob/main/client-library-specfication.md) (Useful for understanding how client libraries are usually structured)

Other than that you can look at the [list of official games](https://github.com/orgs/code-game-project/teams/games/repositories) and try to implement a client for them.
