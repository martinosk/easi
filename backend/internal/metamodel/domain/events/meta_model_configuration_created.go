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
	Pillars   []StrategyPillarData  `json:"pillars"`
	CreatedAt time.Time             `json:"createdAt"`
	CreatedBy string                `json:"createdBy"`
}

func (e MetaModelConfigurationCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

type CreateConfigParams struct {
	ID        string
	TenantID  string
	Sections  []MaturitySectionData
	Pillars   []StrategyPillarData
	CreatedBy string
}

func NewMetaModelConfigurationCreated(params CreateConfigParams) MetaModelConfigurationCreated {
	return MetaModelConfigurationCreated{
		BaseEvent: domain.NewBaseEvent(params.ID),
		ID:        params.ID,
		TenantID:  params.TenantID,
		Sections:  params.Sections,
		Pillars:   params.Pillars,
		CreatedAt: time.Now().UTC(),
		CreatedBy: params.CreatedBy,
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
		"pillars":   e.Pillars,
		"createdAt": e.CreatedAt,
		"createdBy": e.CreatedBy,
	}
}
