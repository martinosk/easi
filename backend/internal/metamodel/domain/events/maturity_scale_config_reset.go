package events

import (
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

type MaturityScaleConfigReset struct {
	domain.BaseEvent
	ID         string                `json:"id"`
	TenantID   string                `json:"tenantId"`
	Version    int                   `json:"version"`
	Sections   []MaturitySectionData `json:"sections"`
	ModifiedAt time.Time             `json:"modifiedAt"`
	ModifiedBy string                `json:"modifiedBy"`
}

func NewMaturityScaleConfigReset(id, tenantID string, version int, sections []MaturitySectionData, modifiedBy string) MaturityScaleConfigReset {
	return MaturityScaleConfigReset{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		TenantID:   tenantID,
		Version:    version,
		Sections:   sections,
		ModifiedAt: time.Now().UTC(),
		ModifiedBy: modifiedBy,
	}
}

func (e MaturityScaleConfigReset) EventType() string {
	return "MaturityScaleConfigReset"
}

func (e MaturityScaleConfigReset) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"tenantId":   e.TenantID,
		"version":    e.Version,
		"sections":   e.Sections,
		"modifiedAt": e.ModifiedAt,
		"modifiedBy": e.ModifiedBy,
	}
}
