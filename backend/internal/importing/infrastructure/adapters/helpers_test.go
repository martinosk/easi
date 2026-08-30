package adapters_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"easi/backend/internal/shared/cqrs"
)

var errDispatchFailed = errors.New("no handler registered")

type recordingBus struct {
	dispatched []cqrs.Command
	createdID  string
	err        error
}

func (b *recordingBus) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	b.dispatched = append(b.dispatched, cmd)
	if b.err != nil {
		return cqrs.EmptyResult(), b.err
	}
	return cqrs.NewResult(b.createdID), nil
}

func (b *recordingBus) Register(string, cqrs.CommandHandler) {}

func failingBus() *recordingBus {
	return &recordingBus{err: errDispatchFailed}
}

func dispatchedCommand[T cqrs.Command](t *testing.T, bus *recordingBus, wantName string) T {
	t.Helper()

	var want T
	if len(bus.dispatched) != 1 {
		t.Fatalf("expected exactly one dispatched command, got %d", len(bus.dispatched))
	}
	cmd, ok := bus.dispatched[0].(T)
	if !ok {
		t.Fatalf("expected dispatched command of type %T, got %T", want, bus.dispatched[0])
	}
	if cmd.CommandName() != wantName {
		t.Errorf("expected command name %q, got %q", wantName, cmd.CommandName())
	}
	return cmd
}

func assertDispatched[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("expected dispatched command %+v, got %+v", want, got)
	}
}

func assertCreatedID(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("expected created id %q, got %q", want, got)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertWrappedDispatchError(t *testing.T, err error, subject string) {
	t.Helper()
	if !errors.Is(err, errDispatchFailed) {
		t.Fatalf("expected error wrapping %v, got %v", errDispatchFailed, err)
	}
	if !strings.Contains(err.Error(), subject) {
		t.Errorf("expected error to mention %q, got %q", subject, err)
	}
}
