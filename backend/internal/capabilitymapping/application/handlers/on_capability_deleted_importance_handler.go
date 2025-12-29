package handlers

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type OnCapabilityDeletedImportanceHandler struct {
	importanceReadModel *readmodels.StrategyImportanceReadModel
}

func NewOnCapabilityDeletedImportanceHandler(importanceReadModel *readmodels.StrategyImportanceReadModel) *OnCapabilityDeletedImportanceHandler {
	return &OnCapabilityDeletedImportanceHandler{importanceReadModel: importanceReadModel}
}

func (h *OnCapabilityDeletedImportanceHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return err
	}

	var capabilityDeleted events.CapabilityDeleted
	if err := json.Unmarshal(eventData, &capabilityDeleted); err != nil {
		return err
	}

	if err := h.importanceReadModel.DeleteByCapability(ctx, capabilityDeleted.ID); err != nil {
		log.Printf("Failed to delete importance ratings for capability %s: %v", capabilityDeleted.ID, err)
		return err
	}

	return nil
}
