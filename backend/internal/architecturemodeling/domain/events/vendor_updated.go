package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type VendorUpdated struct {
	domain.BaseEvent
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	ImplementationPartner string    `json:"implementationPartner"`
	Notes                 string    `json:"notes"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

func (e VendorUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewVendorUpdated(id, name, implementationPartner, notes string) VendorUpdated {
	return VendorUpdated{
		BaseEvent:             domain.NewBaseEvent(id),
		ID:                    id,
		Name:                  name,
		ImplementationPartner: implementationPartner,
		Notes:                 notes,
		UpdatedAt:             time.Now().UTC(),
	}
}

func (e VendorUpdated) EventType() string {
	return "VendorUpdated"
}

func (e VendorUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                    e.ID,
		"name":                  e.Name,
		"implementationPartner": e.ImplementationPartner,
		"notes":                 e.Notes,
		"updatedAt":             e.UpdatedAt,
	}
}
