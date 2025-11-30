package domain

import (
	"context"

	"easi/backend/internal/viewlayouts/domain/aggregates"
	"easi/backend/internal/viewlayouts/domain/valueobjects"
)

type LayoutContainerRepository interface {
	GetByContext(ctx context.Context, contextType valueobjects.LayoutContextType, contextRef valueobjects.ContextRef) (*aggregates.LayoutContainer, error)
	GetByID(ctx context.Context, id valueobjects.LayoutContainerID) (*aggregates.LayoutContainer, error)
	Save(ctx context.Context, container *aggregates.LayoutContainer) error
	Delete(ctx context.Context, id valueobjects.LayoutContainerID) error
	UpsertElementPosition(ctx context.Context, containerID valueobjects.LayoutContainerID, position valueobjects.ElementPosition) error
	DeleteElementPosition(ctx context.Context, containerID valueobjects.LayoutContainerID, elementID valueobjects.ElementID) error
	BatchUpdatePositions(ctx context.Context, containerID valueobjects.LayoutContainerID, positions []valueobjects.ElementPosition) error
	DeleteByContextRef(ctx context.Context, contextType valueobjects.LayoutContextType, contextRef valueobjects.ContextRef) error
	DeleteElementFromAllLayouts(ctx context.Context, elementID valueobjects.ElementID) error
}
