package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

type ApplicationComponentDeleted struct {
	domain.BaseEvent
	ID        string
	Name      string
	DeletedAt time.Time
}

func NewApplicationComponentDeleted(id, name string) ApplicationComponentDeleted {
	return ApplicationComponentDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		Name:      name,
		DeletedAt: time.Now().UTC(),
	}
}

func (e ApplicationComponentDeleted) EventType() string {
	return "ApplicationComponentDeleted"
}

func (e ApplicationComponentDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"name":      e.Name,
		"deletedAt": e.DeletedAt,
	}
}
