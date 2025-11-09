package cqrs

import "context"

// Command represents an intent to change the system state
type Command interface {
	// CommandName returns the name of the command
	CommandName() string
}

// CommandHandler handles a specific command type
type CommandHandler interface {
	// Handle processes the command and returns an error if it fails
	Handle(ctx context.Context, cmd Command) error
}

// CommandBus dispatches commands to their handlers
type CommandBus interface {
	// Dispatch sends a command to its handler
	Dispatch(ctx context.Context, cmd Command) error

	// Register registers a handler for a command type
	Register(commandName string, handler CommandHandler)
}
