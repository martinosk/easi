package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type MaturitySectionData struct {
	Order    int    `json:"order"`
	Name     string `json:"name"`
	MinValue int    `json:"minValue"`
	MaxValue int    `json:"maxValue"`
}

type MetaModelConfigurationCreated struct {
	domain.BaseEvent
	ID        string                `json:"id"`
	TenantID  string                `json:"tenantId"`
	Sections  []MaturitySectionData `json:"sections"`
	CreatedAt time.Time             `json:"createdAt"`
	CreatedBy string                `json:"createdBy"`
}

func NewMetaModelConfigurationCreated(id, tenantID string, sections []MaturitySectionData, createdBy string) MetaModelConfigurationCreated {
	return MetaModelConfigurationCreated{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		TenantID:  tenantID,
		Sections:  sections,
		CreatedAt: time.Now().UTC(),
		CreatedBy: createdBy,
	}
}

func (e MetaModelConfigurationCreated) EventType() string {
	return "MetaModelConfigurationCreated"
}

func (e MetaModelConfigurationCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":        e.ID,
		"tenantId":  e.TenantID,
		"sections":  e.Sections,
		"createdAt": e.CreatedAt,
		"createdBy": e.CreatedBy,
	}
}
