package projectors

import (
	"context"

	"easi/backend/internal/shared/cqrs"
)

type fakeDispatcher struct {
	dispatched []cqrs.Command
}

func (f *fakeDispatcher) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	f.dispatched = append(f.dispatched, cmd)
	return cqrs.EmptyResult(), nil
}
