package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EditGrantRevoked struct {
	domain.BaseEvent
	ID        string    `json:"id"`
	RevokedBy string    `json:"revokedBy"`
	RevokedAt time.Time `json:"revokedAt"`
}

func (e EditGrantRevoked) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewEditGrantRevoked(id, revokedBy string) EditGrantRevoked {
	now := time.Now().UTC()
	return EditGrantRevoked{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		RevokedBy: revokedBy,
		RevokedAt: now,
	}
}

func (e EditGrantRevoked) EventType() string {
	return "EditGrantRevoked"
}

func (e EditGrantRevoked) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"revokedBy": e.RevokedBy,
		"revokedAt": e.RevokedAt.Format(time.RFC3339),
	}
}
