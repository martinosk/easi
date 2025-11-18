package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

type ViewLayoutDirectionUpdated struct {
	domain.BaseEvent
	ViewID          string
	LayoutDirection string
	UpdatedAt       time.Time
}

func NewViewLayoutDirectionUpdated(viewID, layoutDirection string) ViewLayoutDirectionUpdated {
	return ViewLayoutDirectionUpdated{
		BaseEvent:       domain.NewBaseEvent(viewID),
		ViewID:          viewID,
		LayoutDirection: layoutDirection,
		UpdatedAt:       time.Now().UTC(),
	}
}

func (e ViewLayoutDirectionUpdated) EventType() string {
	return "ViewLayoutDirectionUpdated"
}

func (e ViewLayoutDirectionUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"viewId":          e.ViewID,
		"layoutDirection": e.LayoutDirection,
		"updatedAt":       e.UpdatedAt,
	}
}
