package handlers

import (
	"context"
	"errors"
	"fmt"

	"easi/backend/internal/architectureviews/infrastructure/repositories"
	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrDeletedElementWithoutID = errors.New("deletion event carries no element id")

type PlacementRemover interface {
	DeleteElementFromAllViews(ctx context.Context, elementID string, elementType repositories.ElementType) error
}

type DeletedElementHandler struct {
	remover     PlacementRemover
	elementType repositories.ElementType
}

func NewDeletedCapabilityHandler(remover PlacementRemover) *DeletedElementHandler {
	return &DeletedElementHandler{remover: remover, elementType: repositories.ElementTypeCapability}
}

func NewDeletedOriginEntityHandler(remover PlacementRemover) *DeletedElementHandler {
	return &DeletedElementHandler{remover: remover, elementType: repositories.ElementTypeOriginEntity}
}

func (h *DeletedElementHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	elementID := event.AggregateID()
	if elementID == "" {
		return fmt.Errorf("%w: %s", ErrDeletedElementWithoutID, event.EventType())
	}

	if err := h.remover.DeleteElementFromAllViews(ctx, elementID, h.elementType); err != nil {
		return fmt.Errorf("remove deleted %s %s from views: %w", h.elementType, elementID, err)
	}
	return nil
}
