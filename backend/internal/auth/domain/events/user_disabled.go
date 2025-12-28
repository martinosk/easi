package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type UserDisabled struct {
	domain.BaseEvent
	ID         string
	DisabledBy string
	DisabledAt time.Time
}

func NewUserDisabled(id string, disabledBy string) UserDisabled {
	return UserDisabled{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		DisabledBy: disabledBy,
		DisabledAt: time.Now().UTC(),
	}
}

func (e UserDisabled) EventType() string {
	return "UserDisabled"
}

func (e UserDisabled) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"disabledBy": e.DisabledBy,
		"disabledAt": e.DisabledAt.Format(time.RFC3339),
	}
}
