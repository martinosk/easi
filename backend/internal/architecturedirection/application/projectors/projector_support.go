package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/shared/cqrs"
)

func handleProjection[T any](ctx context.Context, eventData []byte, fn func(context.Context, T) error) error {
	var event T
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal %T event data: %w", event, err)
	}
	return fn(ctx, event)
}

type CommandDispatcher interface {
	Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error)
}
