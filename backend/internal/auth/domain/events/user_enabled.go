package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type UserEnabled struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	EnabledBy string    `json:"enabledBy"`
	EnabledAt time.Time `json:"enabledAt"`
}

func (e UserEnabled) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
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
