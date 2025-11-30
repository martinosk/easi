package handlers

import (
	"context"

	"easi/backend/internal/shared/domain"
	viewlayoutsDomain "easi/backend/internal/viewlayouts/domain"
	"easi/backend/internal/viewlayouts/domain/valueobjects"
)

type CapabilityDeletedHandler struct {
	repo viewlayoutsDomain.LayoutContainerRepository
}

func NewCapabilityDeletedHandler(repo viewlayoutsDomain.LayoutContainerRepository) *CapabilityDeletedHandler {
	return &CapabilityDeletedHandler{repo: repo}
}

func (h *CapabilityDeletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData := event.EventData()
	capabilityID, ok := eventData["id"].(string)
	if !ok || capabilityID == "" {
		return nil
	}

	elementID, err := valueobjects.NewElementID(capabilityID)
	if err != nil {
		return nil
	}

	return h.repo.DeleteElementFromAllLayouts(ctx, elementID)
}
