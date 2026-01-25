package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredEntityCreated struct {
	domain.BaseEvent
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	AcquisitionDate   *time.Time `json:"acquisitionDate,omitempty"`
	IntegrationStatus string     `json:"integrationStatus"`
	Notes             string     `json:"notes"`
	CreatedAt         time.Time  `json:"createdAt"`
}

func (e AcquiredEntityCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewAcquiredEntityCreated(id, name string, acquisitionDate *time.Time, integrationStatus, notes string) AcquiredEntityCreated {
	return AcquiredEntityCreated{
		BaseEvent:         domain.NewBaseEvent(id),
		ID:                id,
		Name:              name,
		AcquisitionDate:   acquisitionDate,
		IntegrationStatus: integrationStatus,
		Notes:             notes,
		CreatedAt:         time.Now().UTC(),
	}
}

func (e AcquiredEntityCreated) EventType() string {
	return "AcquiredEntityCreated"
}

func (e AcquiredEntityCreated) EventData() map[string]interface{} {
	data := map[string]interface{}{
		"id":                e.ID,
		"name":              e.Name,
		"integrationStatus": e.IntegrationStatus,
		"notes":             e.Notes,
		"createdAt":         e.CreatedAt,
	}
	if e.AcquisitionDate != nil {
		data["acquisitionDate"] = *e.AcquisitionDate
	}
	return data
}
