package events

import (
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

type MaturityScaleConfigUpdated struct {
	domain.BaseEvent
	ID          string                `json:"id"`
	TenantID    string                `json:"tenantId"`
	Version     int                   `json:"version"`
	NewSections []MaturitySectionData `json:"newSections"`
	ModifiedAt  time.Time             `json:"modifiedAt"`
	ModifiedBy  string                `json:"modifiedBy"`
}

func NewMaturityScaleConfigUpdated(id, tenantID string, version int, newSections []MaturitySectionData, modifiedBy string) MaturityScaleConfigUpdated {
	return MaturityScaleConfigUpdated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		TenantID:    tenantID,
		Version:     version,
		NewSections: newSections,
		ModifiedAt:  time.Now().UTC(),
		ModifiedBy:  modifiedBy,
	}
}

func (e MaturityScaleConfigUpdated) EventType() string {
	return "MaturityScaleConfigUpdated"
}

func (e MaturityScaleConfigUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"tenantId":    e.TenantID,
		"version":     e.Version,
		"newSections": e.NewSections,
		"modifiedAt":  e.ModifiedAt,
		"modifiedBy":  e.ModifiedBy,
	}
}
