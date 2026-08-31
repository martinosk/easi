package cqrs_test

import (
	"context"
	"strings"
	"testing"

	"easi/backend/internal/shared/cqrs"
)

type stubHandler struct{}

func (stubHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	return cqrs.EmptyResult(), nil
}

func TestRegister_DuplicateCommandName_Panics(t *testing.T) {
	bus := cqrs.NewInMemoryCommandBus()
	bus.Register("CreateWidget", stubHandler{})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected Register to panic on duplicate command name, but it did not")
		}
		message, ok := r.(string)
		if !ok {
			t.Fatalf("expected panic value to be a string, got %T: %v", r, r)
		}
		if !strings.Contains(message, "CreateWidget") {
			t.Errorf("expected panic message to name the command %q, got: %s", "CreateWidget", message)
		}
		if !strings.Contains(message, "already registered") {
			t.Errorf("expected panic message to indicate the command is already registered, got: %s", message)
		}
	}()

	bus.Register("CreateWidget", stubHandler{})
}

func TestRegister_DistinctCommandNames_DoesNotPanic(t *testing.T) {
	bus := cqrs.NewInMemoryCommandBus()
	bus.Register("CreateWidget", stubHandler{})
	bus.Register("DeleteWidget", stubHandler{})
}
