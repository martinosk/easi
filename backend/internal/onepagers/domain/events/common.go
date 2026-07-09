package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type SelectionOptionData struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

type FieldRefData struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type ConfigurationEventBase struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	TenantID   string    `json:"tenantId"`
	Version    int       `json:"version"`
	ModifiedAt time.Time `json:"modifiedAt"`
	ModifiedBy string    `json:"modifiedBy"`
}

func (e ConfigurationEventBase) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

type ModifyConfigurationParams struct {
	ConfigID   string
	TenantID   string
	Version    int
	ModifiedBy string
}

func newConfigurationEventBase(params ModifyConfigurationParams) ConfigurationEventBase {
	return ConfigurationEventBase{
		BaseEvent:  domain.NewBaseEvent(params.ConfigID),
		ID:         params.ConfigID,
		TenantID:   params.TenantID,
		Version:    params.Version,
		ModifiedAt: time.Now().UTC(),
		ModifiedBy: params.ModifiedBy,
	}
}

func (e ConfigurationEventBase) baseEventData() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"tenantId":   e.TenantID,
		"version":    e.Version,
		"modifiedAt": e.ModifiedAt,
		"modifiedBy": e.ModifiedBy,
	}
}
