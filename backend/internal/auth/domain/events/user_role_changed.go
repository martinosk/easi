package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type UserRoleChanged struct {
	domain.BaseEvent
	ID          string
	OldRole     string
	NewRole     string
	ChangedByID string
	ChangedAt   time.Time
}

func NewUserRoleChanged(
	id string,
	oldRole string,
	newRole string,
	changedByID string,
) UserRoleChanged {
	return UserRoleChanged{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		OldRole:     oldRole,
		NewRole:     newRole,
		ChangedByID: changedByID,
		ChangedAt:   time.Now().UTC(),
	}
}

func (e UserRoleChanged) EventType() string {
	return "UserRoleChanged"
}

func (e UserRoleChanged) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"oldRole":     e.OldRole,
		"newRole":     e.NewRole,
		"changedById": e.ChangedByID,
		"changedAt":   e.ChangedAt.Format(time.RFC3339),
	}
}
