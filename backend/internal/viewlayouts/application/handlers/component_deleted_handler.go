package handlers

import (
	"context"

	"easi/backend/internal/shared/domain"
	viewlayoutsDomain "easi/backend/internal/viewlayouts/domain"
	"easi/backend/internal/viewlayouts/domain/valueobjects"
)

type ComponentDeletedHandler struct {
	repo viewlayoutsDomain.LayoutContainerRepository
}

func NewComponentDeletedHandler(repo viewlayoutsDomain.LayoutContainerRepository) *ComponentDeletedHandler {
	return &ComponentDeletedHandler{repo: repo}
}

func (h *ComponentDeletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData := event.EventData()
	componentID, ok := eventData["id"].(string)
	if !ok || componentID == "" {
		return nil
	}

	elementID, err := valueobjects.NewElementID(componentID)
	if err != nil {
		return nil
	}

	return h.repo.DeleteElementFromAllLayouts(ctx, elementID)
}
