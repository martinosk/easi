package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type ViewEdgeTypeUpdated struct {
	domain.BaseEvent
	ViewID    string
	EdgeType  string
	UpdatedAt time.Time
}

func NewViewEdgeTypeUpdated(viewID, edgeType string) ViewEdgeTypeUpdated {
	return ViewEdgeTypeUpdated{
		BaseEvent: domain.NewBaseEvent(viewID),
		ViewID:    viewID,
		EdgeType:  edgeType,
		UpdatedAt: time.Now().UTC(),
	}
}

func (e ViewEdgeTypeUpdated) EventType() string {
	return "ViewEdgeTypeUpdated"
}

func (e ViewEdgeTypeUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"viewId":    e.ViewID,
		"edgeType":  e.EdgeType,
		"updatedAt": e.UpdatedAt,
	}
}
