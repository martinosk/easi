package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type OnePagerConfigurationCreated struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	SubjectType string    `json:"subjectType"`
	BuiltIns    []string  `json:"builtIns"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
}

type CreateConfigurationParams struct {
	ID          string
	TenantID    string
	SubjectType string
	BuiltIns    []string
	CreatedBy   string
}

func NewOnePagerConfigurationCreated(params CreateConfigurationParams) OnePagerConfigurationCreated {
	return OnePagerConfigurationCreated{
		BaseEvent:   domain.NewBaseEvent(params.ID),
		ID:          params.ID,
		TenantID:    params.TenantID,
		SubjectType: params.SubjectType,
		BuiltIns:    params.BuiltIns,
		CreatedAt:   time.Now().UTC(),
		CreatedBy:   params.CreatedBy,
	}
}

func (e OnePagerConfigurationCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e OnePagerConfigurationCreated) EventType() string {
	return "OnePagerConfigurationCreated"
}

func (e OnePagerConfigurationCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"tenantId":    e.TenantID,
		"subjectType": e.SubjectType,
		"builtIns":    e.BuiltIns,
		"createdAt":   e.CreatedAt,
		"createdBy":   e.CreatedBy,
	}
}
