package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

// ViewDeleted is raised when an architecture view is deleted
type ViewDeleted struct {
	domain.BaseEvent
	ViewID    string
	DeletedAt time.Time
}

// NewViewDeleted creates a new ViewDeleted event
func NewViewDeleted(viewID string) ViewDeleted {
	return ViewDeleted{
		BaseEvent: domain.NewBaseEvent(viewID),
		ViewID:    viewID,
		DeletedAt: time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ViewDeleted) EventType() string {
	return "ViewDeleted"
}

// EventData returns the event data as a map for serialization
func (e ViewDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"viewId":    e.ViewID,
		"deletedAt": e.DeletedAt,
	}
}
