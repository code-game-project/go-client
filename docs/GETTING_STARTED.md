# Getting Started

This guide will walk you through creating your own client for the [tic-tac-toe-simple](https://github.com/code-game-project/tic-tac-toe-simple)
game at `games.code-game.org/tic-tac-toe-simple`.
It is recommended that you first read the general CodeGame [getting started guide](https://docs.code-game.org/guides/getting-started).

## Scope

This guide will teach you the most important functions of this client library.
It does not focus on creating a beautiful and pleasant to use application. This task is up to you.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Setting up the project](#setting-up-the-project)
- [The main function](#the-main-function)
  - [CLI arguments](#cli-arguments)
  - [Joining a game](#joining-a-game)
  - [The event loop](#the-event-loop)
- [Listening for events](#listening-for-events)
- [Sending events](#sending-events)
- [Implementing the tic-tac-toe-simple client](#implementing-the-tic-tac-toe-simple-client)
  - [The client struct](#the-client-struct)
  - [Handling events](#handling-events)
    - [The start event](#the-start-event)
    - [The invalid_action event](#the-invalid_action-event)
    - [The board event](#the-board-event)
    - [The turn event](#the-turn-event)
    - [The mark event](#the-mark-event)
    - [The game_over event](#the-game_over-event)
- [Running the game](#running-the-game)
- [What next?](#what-next)
- [Complete main.go](#complete-maingo)

## Prerequisites

In order to follow this guide you will have to have the following software installed:

- [Go](https://go.dev) 1.18+
- [CodeGame CLI](https://github.com/code-game-project/codegame-cli)

## Setting up the project

The [CodeGame CLI](https://github.com/code-game-project/codegame-cli) allows you to quickly get started with writing CodeGame applications.

To create a new game client simply execute `codegame new` in a terminal, choose 'Game Client' as the project type and enter these values:
- Project name: *tictactoe-client*
- Game server URL: *games.code-game.org/tic-tac-toe-simple*
- Language: *Go*
- Project module path: *tictactoe-client*

Finally choose whether you want to initialize Git, create a README or create a LICENSE.

The project will be available at `tictactoe-client/`. All of the code in this guide will be written into the `main.go` file.

**NOTE:** *codegame-cli* actually creates a wrapper around this library to make some tasks easier. We will be using these wrappers in the following sections instead of the bare library.
If you want to learn how the library works without wrappers, you can view generated files in `tictactoesimple/`.

## The main function

*codegame-cli* already created some boilerplate code to create new games and join existing ones.

The `main` function starts by declaring some command line flags with [pflag](https://github.com/spf13/pflag), a drop-in replacement for Go's flag package, which implements POSIX/GNU-style --flags:

### CLI arguments
```go
var create bool
pflag.BoolVarP(&create, "create", "c", false, "Create a new game.")
var public bool
pflag.BoolVarP(&public, "public", "p", false, "Make the created game public.")
var gameId string
pflag.StringVarP(&gameId, "join", "j", "", "Join a game.")
pflag.Parse()

if pflag.NArg() != 1 {
	fmt.Fprintf(os.Stderr, "USAGE: %s [OPTIONS] <username>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "For help use --help.\n")
	os.Exit(1)
}
username := pflag.Arg(0)
```

These flags allow the user to choose whether they want to create a new game with `tictactoe-client --create <username>`,
join an existing one with `tictactoe-client --join=<game_id> <username>` or reconnect to a previous session with `tictactoe-client <username>`.

### Joining a game

Depending on the flags chosen the next block of code will create a new game, join an existing game or reconnect to a previous session, and prints the ID of the joined game:

```go
if create {
	game, err = tictactoesimple.CreateAndJoinGame(public, username)
} else if gameId != "" {
	game, err = tictactoesimple.JoinGame(gameId, username)
} else {
	game, err = tictactoesimple.ReconnectGame(username)
}

if err != nil {
	fmt.Println("Couldn't join game:", err)
	os.Exit(1)
}

fmt.Println("GameID:", game.Id)
```

### The event loop

At the end of the `main` function a call to `game.Run` will enter a listening loop, which will receive the events from the server and call registered event listeners.

Alternatively the `game.Update` function can be called in a loop. This method of polling events is most often used with frameworks that need their own game loop like Unity or SDL.

## Listening for events

You can register an event listener with the `game.On…` or the `game.On…Once` methods.

For example to register event handler for a `hello_world` event you would use the following snippet:

```go
game.OnHelloWorldEvent(func(origin mygame.Player, data cg.HelloWorldEventData) {
	fmt.Println("Hello, World!")
})
```

## Sending events

Events can be sent with the `game.Send…` methods.

To send a `hello_world` event you would use this block of code:

```go
game.SendHelloWorldEvent(tictactoesimple.HelloWorldEventData{})
```

## Implementing the tic-tac-toe-simple client

Now that we have the basics out of the way we can begin writing the client for *tic-tac-toe-simple*.

### The client struct

Because we will need the `game` variable throughout the application it is advised to implement all of the game logic as methods of a `client` struct which stores the game and other useful state.

Let's declare and use the `client` struct:

```go
// package ...
// import (...)

// The client struct stores the game and the sign of the current player ('x' or 'o')
type client struct {
	game *tictactoesimple.Game

	sign tictactoesimple.Sign
}

// Instead of calling `game.Run` directly in the main function we will call it in `client.run` after registering all needed event listeners.
func (c *client) run() error {
	// TODO: Register event listeners.

	return c.game.Run()
}

func main() {
	// ...

	// Replace `game.Run()` with:
	client := &client{
		game: game,
	}
	client.run()
}
```

### Handling events

#### The `start` event

Once a match is found the server sends the `start` event which includes the player IDs mapped to their signs.

The `start` event is only sent once. Because the listener won't be needed anymore after receiving the event, we use the `OnStartEventOnce` method.

```go
func (c *client) run() error {
	c.game.OnStartEventOnce(func(origin tictactoesimple.Player, data tictactoesimple.StartEventData) {
		// `game.Session` returns a struct with useful information like the current game ID or the player ID.
		// In this case we need the player ID to receive the sign of our player.
		c.sign = data.Signs[c.game.Session().PlayerId]

		// Print the sign.
		fmt.Println("Found a match! Your sign is:", c.sign)
	})

	// return c.game.Run()
}
```

#### The `invalid_action` event

Another useful event is the `invalid_action` event which is sent when the player does something wrong like trying to mark an already occupied field.
In this case we want to print the error message.

```go
func (c *client) run() error {
	// c.game.OnStartEventOnce(func(origin tictactoesimple.Player, data tictactoesimple.StartEventData) {...}

	c.game.OnInvalidActionEvent(func(origin tictactoesimple.Player, data tictactoesimple.InvalidActionEventData) {
		// Print the error message.
		fmt.Println(data.Message)
	})

	// return c.game.Run()
}
```

#### The `board` event

The `board` event tells us the current state of the board.
Every time we receive this event we want to print the board to the console.

```go
func (c *client) run() error {
	// c.game.OnInvalidActionEvent(func(origin tictactoesimple.Player, data tictactoesimple.InvalidActionEventData) {...}

	c.game.OnBoardEvent(func(origin tictactoesimple.Player, data tictactoesimple.BoardEventData) {
		c.printBoard(data.Board)
	})

	// return c.game.Run()
}

func (c *client) printBoard(board [][]tictactoesimple.Field) {
	// Print a separator.
	fmt.Println(strings.Repeat("=", 50))

	// Loop through all rows.
	for i := range board {
		// Loop through all columns.
		for j := range board[i] {
			// Print a '/' symbol if the field is not occupied.
			if board[i][j].Sign == tictactoesimple.SignNone {
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

#### The `turn` event

What good is a board if we can't use it? Well, the `turn` event notifies us which player's turn it currently is so they can mark a field.

Once we receive a `turn` event we need to check whether it's our turn and let the player input a field if it is.

```go
func (c *client) run() error {
	// c.game.OnBoardEvent(func(origin tictactoesimple.Player, data tictactoesimple.BoardEventData) {

	c.game.OnTurnEvent(func(origin tictactoesimple.Player, data tictactoesimple.TurnEventData) {
		if data.Sign == c.sign {
			// It's our turn.
			fmt.Println(strings.Repeat("=", 50))
			c.mark()
		} else {
			// It's our opponent's turn.
			fmt.Println("Waiting for opponent…")
		}
	})

	// return c.game.Run()
}

func (c *client) mark() {
	// TODO
}
```

#### The `mark` event

There is only one event, we need send to the server: the `mark` event. It allows us to mark an empty field with our sign provided it's our turn.

We already know when it's our turn and call the `mark` method so let's let the user input a field and send it to the server.

```go
func (c *client) mark() {
	// Ask the user to input a field (e.g. 1,1 for the top left field)
	fmt.Print("Where do you want to place your sign? (row,column) ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	location := scanner.Text()

	// TODO: input validation

	// Split row and column.
	coords := strings.Split(location, ",")

	// Convert row and column to an integer.
	row, _ := strconv.Atoi(coords[0])
	column, _ := strconv.Atoi(coords[1])

	// Send the `mark` event with the row and column to the server.
	c.game.SendMarkEvent(tictactoesimple.MarkEventData{
		// Subtract 1 because the user enters a 1 based row number (1,2,3) while the server accepts a 0 based row number (0,1,2).
		Row:    row - 1,
		Column: column - 1,
	})
}
```

We also probably want to let the user enter again if they made a mistake.
Because of that we call the `mark` method at the bottom of the `invalid_action` event callback:

```go
func (c *client) run() error {
	// c.game.OnStartEventOnce(func(origin tictactoesimple.Player, data tictactoesimple.StartEventData) {...}

	c.game.OnInvalidActionEvent(func(origin tictactoesimple.Player, data tictactoesimple.InvalidActionEventData) {
		// Print the error message.
		fmt.Println(data.Message)

		c.mark() // <-------
	})

	// c.game.OnBoardEvent(func(origin tictactoesimple.Player, data tictactoesimple.BoardEventData) {
}
```

#### The `game_over` event

There is only one event left to go. The `game_over` event is sent once either all fields have been marked or a player has won.

Apart from the type of ending and the winning sign the `game_over` event also returns the fields which form the winning row.
For simplicity we will only print the outcome.

```go
	// c.game.OnTurnEvent(func(origin tictactoesimple.Player, data tictactoesimple.TurnEventData) {...}

	c.game.OnGameOverEvent(func(origin tictactoesimple.Player, data tictactoesimple.GameOverEventData) {
		fmt.Println(strings.Repeat("=", 50))

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

	// return c.game.Run()
```

## Running the game

With only about 160 lines of code our tic-tac-toe-simple client is finished.
Let's try it out!

First open a terminal in the project directory and create a new game:
```sh
go run . --create user1
```

This will print something similar to this:
```
GameID: d696ce79-b093-46df-9e71-bd15f86e58a6
```

You can now use the game ID to join the game in a second terminal window:
```sh
go run . --join=d696ce79-b093-46df-9e71-bd15f86e58a6 user2
```

Now enjoy your very own tic-tac-toe multiplayer game.

## What next?

I recommend reading the following specifications to build a stronger understanding of CodeGame:

- [CodeGame Protocol Specification](https://docs.code-game.org/specifications/protocol) (Useful if you want to know how CodeGame works under the hood)
- [CodeGame Events Language Specification](https://docs.code-game.org/specifications/cge) (Useful if you want to write a game or understand existing games better)
- [CodeGame Game Server Specification](https://docs.code-game.org/specifications/game-server) (Mostly useful if you plan to write a client which displays more information like a list of public games from a server)
- [CodeGame Client Library Specification](https://docs.code-game.org/specifications/client-library) (Useful for understanding how client libraries are usually structured)

Other than that you can look at the [list of official games](https://games.code-game.org) and try to implement a client for them.

## Complete main.go

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"tictactoe-client/tictactoesimple"

	"github.com/code-game-project/go-client/cg"
	"github.com/spf13/pflag"
)

type client struct {
	game *tictactoesimple.Game

	// The sign of the current player.
	sign tictactoesimple.Sign
}

func (c *client) run() error {
	c.game.OnStartEventOnce(func(origin tictactoesimple.Player, data tictactoesimple.StartEventData) {
		// `game.Session` returns a struct with useful information like the current game ID or the player ID.
		// In this case we need the player ID to receive the sign of our player.
		c.sign = data.Signs[c.game.Session().PlayerId]

		// Print the sign.
		fmt.Println("Found a match! Your sign is:", c.sign)
	})

	c.game.OnInvalidActionEvent(func(origin tictactoesimple.Player, data tictactoesimple.InvalidActionEventData) {
		// Print the error message.
		fmt.Println(data.Message)

		c.mark()
	})

	c.game.OnBoardEvent(func(origin tictactoesimple.Player, data tictactoesimple.BoardEventData) {
		c.printBoard(data.Board)
	})

	c.game.OnTurnEvent(func(origin tictactoesimple.Player, data tictactoesimple.TurnEventData) {
		if data.Sign == c.sign {
			// It's our turn.
			fmt.Println(strings.Repeat("=", 50))
			c.mark()
		} else {
			// It's our opponent's turn.
			fmt.Println("Waiting for opponent…")
		}
	})

	c.game.OnGameOverEvent(func(origin tictactoesimple.Player, data tictactoesimple.GameOverEventData) {
		fmt.Println(strings.Repeat("=", 50))

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

	return c.game.Run()
}

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
	c.game.SendMarkEvent(tictactoesimple.MarkEventData{
		// Subtract 1 because the user enters a 1 based row number (1,2,3) while the server accepts a 0 based row number (0,1,2).
		Row:    row - 1,
		Column: column - 1,
	})
}

func (c *client) printBoard(board [][]tictactoesimple.Field) {
	// Print a separator.
	fmt.Println(strings.Repeat("=", 50))

	// Loop through all rows.
	for i := range board {
		// Loop through all columns.
		for j := range board[i] {
			// Print a '/' symbol if the field is not occupied.
			if board[i][j].Sign == tictactoesimple.SignNone {
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

func main() {
	var create bool
	pflag.BoolVarP(&create, "create", "c", false, "Create a new game.")
	var public bool
	pflag.BoolVarP(&public, "public", "p", false, "Make the created game public.")
	var gameId string
	pflag.StringVarP(&gameId, "join", "j", "", "Join a game.")
	pflag.Parse()

	if pflag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "USAGE: %s [OPTIONS] <username>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "For help use --help.\n")
		os.Exit(1)
	}
	username := pflag.Arg(0)

	var game *tictactoesimple.Game
	var err error

	if create {
		game, err = tictactoesimple.CreateAndJoinGame(public, username)
	} else if gameId != "" {
		game, err = tictactoesimple.JoinGame(gameId, username)
	} else {
		game, err = tictactoesimple.ReconnectGame(username)
	}

	if err != nil {
		fmt.Println("Couldn't join game:", err)
		os.Exit(1)
	}

	fmt.Println("GameID:", game.Id)

	game.OnCGErrorEvent(func(origin tictactoesimple.Player, data cg.ErrorEventData) {
		fmt.Println("error:", data.Message)
	})

	client := &client{
		game: game,
	}
	client.run()
}
```
