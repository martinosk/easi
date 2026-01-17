package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type VendorDeleted struct {
	domain.BaseEvent
	ID        string
	Name      string
	DeletedAt time.Time
}

func NewVendorDeleted(id, name string) VendorDeleted {
	return VendorDeleted{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		Name:      name,
		DeletedAt: time.Now().UTC(),
	}
}

func (e VendorDeleted) EventType() string {
	return "VendorDeleted"
}

func (e VendorDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"name":      e.Name,
		"deletedAt": e.DeletedAt,
	}
}
