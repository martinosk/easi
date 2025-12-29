package handlers

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type OnBusinessDomainDeletedImportanceHandler struct {
	importanceReadModel *readmodels.StrategyImportanceReadModel
}

func NewOnBusinessDomainDeletedImportanceHandler(importanceReadModel *readmodels.StrategyImportanceReadModel) *OnBusinessDomainDeletedImportanceHandler {
	return &OnBusinessDomainDeletedImportanceHandler{importanceReadModel: importanceReadModel}
}

func (h *OnBusinessDomainDeletedImportanceHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return err
	}

	var domainDeleted events.BusinessDomainDeleted
	if err := json.Unmarshal(eventData, &domainDeleted); err != nil {
		return err
	}

	if err := h.importanceReadModel.DeleteByBusinessDomain(ctx, domainDeleted.ID); err != nil {
		log.Printf("Failed to delete importance ratings for domain %s: %v", domainDeleted.ID, err)
		return err
	}

	return nil
}
