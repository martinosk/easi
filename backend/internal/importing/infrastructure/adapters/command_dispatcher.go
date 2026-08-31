package adapters

import (
	"context"
	"fmt"

	"easi/backend/internal/shared/cqrs"
)

type commandDispatcher struct {
	commandBus cqrs.CommandBus
}

func (d commandDispatcher) createdID(ctx context.Context, cmd cqrs.Command, subject string) (string, error) {
	result, err := d.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("dispatch %s command for %s: %w", cmd.CommandName(), subject, err)
	}
	return result.CreatedID, nil
}

func (d commandDispatcher) dispatch(ctx context.Context, cmd cqrs.Command, subject string) error {
	_, err := d.createdID(ctx, cmd, subject)
	return err
}
