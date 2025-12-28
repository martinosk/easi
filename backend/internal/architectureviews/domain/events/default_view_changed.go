package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

// DefaultViewChanged is raised when the default view is changed
type DefaultViewChanged struct {
	domain.BaseEvent
	ViewID    string
	IsDefault bool
	ChangedAt time.Time
}

// NewDefaultViewChanged creates a new DefaultViewChanged event
func NewDefaultViewChanged(viewID string, isDefault bool) DefaultViewChanged {
	return DefaultViewChanged{
		BaseEvent: domain.NewBaseEvent(viewID),
		ViewID:    viewID,
		IsDefault: isDefault,
		ChangedAt: time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e DefaultViewChanged) EventType() string {
	return "DefaultViewChanged"
}

// EventData returns the event data as a map for serialization
func (e DefaultViewChanged) EventData() map[string]interface{} {
	return map[string]interface{}{
		"viewId":    e.ViewID,
		"isDefault": e.IsDefault,
		"changedAt": e.ChangedAt,
	}
}
