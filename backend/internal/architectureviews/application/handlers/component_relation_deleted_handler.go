package handlers

import (
	"context"
	domain "easi/backend/internal/shared/eventsourcing"
	"log"
)

type ComponentRelationDeletedHandler struct{}

func NewComponentRelationDeletedHandler() *ComponentRelationDeletedHandler {
	return &ComponentRelationDeletedHandler{}
}

func (h *ComponentRelationDeletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	relationID := event.AggregateID()

	log.Printf("Handling ComponentRelationDeleted for relation %s", relationID)

	return nil
}
