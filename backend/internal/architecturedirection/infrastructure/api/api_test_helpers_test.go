package api

import (
	"context"

	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type mockCommandBus struct {
	dispatched    []cqrs.Command
	createdID     string
	err           error
	afterDispatch func()
}

func (m *mockCommandBus) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	if m.err != nil {
		return cqrs.EmptyResult(), m.err
	}
	m.dispatched = append(m.dispatched, cmd)
	if m.afterDispatch != nil {
		m.afterDispatch()
	}
	return cqrs.CommandResult{CreatedID: m.createdID}, nil
}

func (m *mockCommandBus) Register(_ string, _ cqrs.CommandHandler) {}

func architectActor() sharedctx.Actor {
	return sharedctx.NewActor("u1", "user@example.com", sharedctx.RoleArchitect)
}

func stakeholderActor() sharedctx.Actor {
	return sharedctx.NewActor("u2", "stake@example.com", sharedctx.RoleStakeholder)
}
