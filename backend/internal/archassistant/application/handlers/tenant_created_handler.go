package handlers

import (
	"context"
	"log"

	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	sharedctx "easi/backend/internal/shared/context"
	domain2 "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type TenantCreatedHandler struct {
	repo domain.AIConfigurationRepository
}

func NewTenantCreatedHandler(repo domain.AIConfigurationRepository) *TenantCreatedHandler {
	return &TenantCreatedHandler{repo: repo}
}

func (h *TenantCreatedHandler) Handle(ctx context.Context, event domain2.DomainEvent) error {
	tenantID := event.AggregateID()

	log.Printf("Provisioning AI configuration for tenant %s", tenantID)

	tenantVo, err := sharedvo.NewTenantID(tenantID)
	if err != nil {
		log.Printf("Error creating tenant ID for AI config provisioning: %v", err)
		return err
	}
	ctx = sharedctx.WithTenant(ctx, tenantVo)

	config := aggregates.NewAIConfiguration(tenantID)
	if err := h.repo.Save(ctx, config); err != nil {
		log.Printf("Error provisioning AI configuration for tenant %s: %v", tenantID, err)
		return err
	}

	log.Printf("AI configuration provisioned for tenant %s with ID %s", tenantID, config.ID())
	return nil
}
