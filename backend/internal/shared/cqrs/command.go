package cqrs

import "context"

// Command represents an intent to change the system state
type Command interface {
	// CommandName returns the name of the command
	CommandName() string
}

// CommandResult contains the result of a command execution
type CommandResult struct {
	// CreatedID is the ID of the newly created aggregate (for create commands)
	CreatedID string
}

// EmptyResult returns an empty command result (for commands that don't create new aggregates)
func EmptyResult() CommandResult {
	return CommandResult{}
}

// NewResult creates a result with the created aggregate ID
func NewResult(createdID string) CommandResult {
	return CommandResult{CreatedID: createdID}
}

// CommandHandler handles a specific command type
type CommandHandler interface {
	// Handle processes the command and returns a result and error
	Handle(ctx context.Context, cmd Command) (CommandResult, error)
}

// CommandBus dispatches commands to their handlers
type CommandBus interface {
	// Dispatch sends a command to its handler and returns the result
	Dispatch(ctx context.Context, cmd Command) (CommandResult, error)

	// Register registers a handler for a command type
	Register(commandName string, handler CommandHandler)
}
