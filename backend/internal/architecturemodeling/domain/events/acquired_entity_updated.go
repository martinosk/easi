package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredEntityUpdated struct {
	domain.BaseEvent
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	AcquisitionDate   *time.Time `json:"acquisitionDate,omitempty"`
	IntegrationStatus string     `json:"integrationStatus"`
	Notes             string     `json:"notes"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

func (e AcquiredEntityUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewAcquiredEntityUpdated(id, name string, acquisitionDate *time.Time, integrationStatus, notes string) AcquiredEntityUpdated {
	return AcquiredEntityUpdated{
		BaseEvent:         domain.NewBaseEvent(id),
		ID:                id,
		Name:              name,
		AcquisitionDate:   acquisitionDate,
		IntegrationStatus: integrationStatus,
		Notes:             notes,
		UpdatedAt:         time.Now().UTC(),
	}
}

func (e AcquiredEntityUpdated) EventType() string {
	return "AcquiredEntityUpdated"
}

func (e AcquiredEntityUpdated) EventData() map[string]interface{} {
	data := map[string]interface{}{
		"id":                e.ID,
		"name":              e.Name,
		"integrationStatus": e.IntegrationStatus,
		"notes":             e.Notes,
		"updatedAt":         e.UpdatedAt,
	}
	if e.AcquisitionDate != nil {
		data["acquisitionDate"] = *e.AcquisitionDate
	}
	return data
}
