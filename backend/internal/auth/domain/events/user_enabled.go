package events

import (
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

type UserEnabled struct {
	domain.BaseEvent
	ID        string
	EnabledBy string
	EnabledAt time.Time
}

func NewUserEnabled(id string, enabledBy string) UserEnabled {
	return UserEnabled{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		EnabledBy: enabledBy,
		EnabledAt: time.Now().UTC(),
	}
}

func (e UserEnabled) EventType() string {
	return "UserEnabled"
}

func (e UserEnabled) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"enabledBy": e.EnabledBy,
		"enabledAt": e.EnabledAt.Format(time.RFC3339),
	}
}
