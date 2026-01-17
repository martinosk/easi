package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredEntityUpdated struct {
	domain.BaseEvent
	ID                string
	Name              string
	AcquisitionDate   *time.Time
	IntegrationStatus string
	Notes             string
	UpdatedAt         time.Time
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
