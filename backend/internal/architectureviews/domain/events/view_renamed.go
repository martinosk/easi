package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

// ViewRenamed is raised when an architecture view is renamed
type ViewRenamed struct {
	domain.BaseEvent
	ViewID    string    `json:"viewId"`
	OldName   string    `json:"oldName"`
	NewName   string    `json:"newName"`
	RenamedAt time.Time `json:"renamedAt"`
}

func (e ViewRenamed) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ViewID
}

// NewViewRenamed creates a new ViewRenamed event
func NewViewRenamed(viewID, oldName, newName string) ViewRenamed {
	return ViewRenamed{
		BaseEvent: domain.NewBaseEvent(viewID),
		ViewID:    viewID,
		OldName:   oldName,
		NewName:   newName,
		RenamedAt: time.Now().UTC(),
	}
}

// EventType returns the event type name
func (e ViewRenamed) EventType() string {
	return "ViewRenamed"
}

// EventData returns the event data as a map for serialization
func (e ViewRenamed) EventData() map[string]interface{} {
	return map[string]interface{}{
		"viewId":    e.ViewID,
		"oldName":   e.OldName,
		"newName":   e.NewName,
		"renamedAt": e.RenamedAt,
	}
}
