package handlers

import (
	"context"
	"log"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type TenantCreatedHandler struct {
	commandBus cqrs.CommandBus
}

func NewTenantCreatedHandler(commandBus cqrs.CommandBus) *TenantCreatedHandler {
	return &TenantCreatedHandler{
		commandBus: commandBus,
	}
}

func (h *TenantCreatedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	tenantID := event.AggregateID()
	data := event.EventData()

	firstAdminEmail, _ := data["firstAdminEmail"].(string)
	if firstAdminEmail == "" {
		firstAdminEmail = "system@easi.io"
	}

	log.Printf("Provisioning MetaModel configuration for tenant %s", tenantID)

	cmd := &commands.CreateMetaModelConfiguration{
		TenantID:  tenantID,
		CreatedBy: firstAdminEmail,
	}

	result, err := h.commandBus.Dispatch(ctx, cmd)
	if err != nil {
		log.Printf("Error provisioning MetaModel configuration for tenant %s: %v", tenantID, err)
		return err
	}

	log.Printf("MetaModel configuration provisioned for tenant %s with ID %s", tenantID, result.CreatedID)

	return nil
}
