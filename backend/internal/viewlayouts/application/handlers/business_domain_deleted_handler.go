package handlers

import (
	"context"

	domain "easi/backend/internal/shared/eventsourcing"
	viewlayoutsDomain "easi/backend/internal/viewlayouts/domain"
	"easi/backend/internal/viewlayouts/domain/valueobjects"
)

type BusinessDomainDeletedHandler struct {
	repo viewlayoutsDomain.LayoutContainerRepository
}

func NewBusinessDomainDeletedHandler(repo viewlayoutsDomain.LayoutContainerRepository) *BusinessDomainDeletedHandler {
	return &BusinessDomainDeletedHandler{repo: repo}
}

func (h *BusinessDomainDeletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData := event.EventData()
	domainID, ok := eventData["id"].(string)
	if !ok || domainID == "" {
		return nil
	}

	contextRef, err := valueobjects.NewContextRef(domainID)
	if err != nil {
		return nil
	}

	return h.repo.DeleteByContextRef(ctx, valueobjects.ContextTypeBusinessDomainGrid, contextRef)
}
