package handlers

import (
	"context"
	"fmt"
	"strings"

	"easi/backend/internal/auth/application/commands"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

const firstAdminRole = "admin"

type CommandDispatcher interface {
	Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error)
}

type TenantCreatedReactor struct {
	commandBus CommandDispatcher
}

func NewTenantCreatedReactor(commandBus CommandDispatcher) *TenantCreatedReactor {
	return &TenantCreatedReactor{commandBus: commandBus}
}

func (r *TenantCreatedReactor) Handle(ctx context.Context, event domain.DomainEvent) error {
	email, _ := event.EventData()["firstAdminEmail"].(string)
	email = strings.TrimSpace(email)
	if email == "" {
		return nil
	}

	tenantID, err := sharedvo.NewTenantID(event.AggregateID())
	if err != nil {
		return fmt.Errorf("read tenant %s from %s: %w", event.AggregateID(), event.EventType(), err)
	}

	cmd := &commands.CreateInvitation{Email: email, Role: firstAdminRole}
	if _, err := r.commandBus.Dispatch(sharedctx.WithTenant(ctx, tenantID), cmd); err != nil {
		return fmt.Errorf("invite first admin of tenant %s: %w", tenantID.Value(), err)
	}
	return nil
}
