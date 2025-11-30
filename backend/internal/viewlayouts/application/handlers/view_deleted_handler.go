package handlers

import (
	"context"

	"easi/backend/internal/shared/domain"
	viewlayoutsDomain "easi/backend/internal/viewlayouts/domain"
	"easi/backend/internal/viewlayouts/domain/valueobjects"
)

type ViewDeletedHandler struct {
	repo viewlayoutsDomain.LayoutContainerRepository
}

func NewViewDeletedHandler(repo viewlayoutsDomain.LayoutContainerRepository) *ViewDeletedHandler {
	return &ViewDeletedHandler{repo: repo}
}

func (h *ViewDeletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData := event.EventData()
	viewID, ok := eventData["id"].(string)
	if !ok || viewID == "" {
		return nil
	}

	contextRef, err := valueobjects.NewContextRef(viewID)
	if err != nil {
		return nil
	}

	return h.repo.DeleteByContextRef(ctx, valueobjects.ContextTypeArchitectureCanvas, contextRef)
}
