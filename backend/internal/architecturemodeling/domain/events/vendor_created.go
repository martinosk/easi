package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type VendorCreated struct {
	domain.BaseEvent
	ID                    string
	Name                  string
	ImplementationPartner string
	Notes                 string
	CreatedAt             time.Time
}

func NewVendorCreated(id, name, implementationPartner, notes string) VendorCreated {
	return VendorCreated{
		BaseEvent:             domain.NewBaseEvent(id),
		ID:                    id,
		Name:                  name,
		ImplementationPartner: implementationPartner,
		Notes:                 notes,
		CreatedAt:             time.Now().UTC(),
	}
}

func (e VendorCreated) EventType() string {
	return "VendorCreated"
}

func (e VendorCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                    e.ID,
		"name":                  e.Name,
		"implementationPartner": e.ImplementationPartner,
		"notes":                 e.Notes,
		"createdAt":             e.CreatedAt,
	}
}
