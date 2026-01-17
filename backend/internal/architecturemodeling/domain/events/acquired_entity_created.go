package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredEntityCreated struct {
	domain.BaseEvent
	ID                string
	Name              string
	AcquisitionDate   *time.Time
	IntegrationStatus string
	Notes             string
	CreatedAt         time.Time
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
