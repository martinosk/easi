package events

import (
	"time"

	"easi/backend/internal/onepagers/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type FactsEventBase struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	SubjectType string    `json:"subjectType"`
	SubjectID   string    `json:"subjectId"`
	Version     int       `json:"version"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	ModifiedBy  string    `json:"modifiedBy"`
}

func (e FactsEventBase) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

type ModifyFactsParams struct {
	FactsID     string
	TenantID    string
	SubjectType string
	SubjectID   string
	Version     int
	ModifiedBy  string
}

func newFactsEventBase(params ModifyFactsParams) FactsEventBase {
	return FactsEventBase{
		BaseEvent:   domain.NewBaseEvent(params.FactsID),
		ID:          params.FactsID,
		TenantID:    params.TenantID,
		SubjectType: params.SubjectType,
		SubjectID:   params.SubjectID,
		Version:     params.Version,
		ModifiedAt:  time.Now().UTC(),
		ModifiedBy:  params.ModifiedBy,
	}
}

func (e FactsEventBase) baseFactsEventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"tenantId":    e.TenantID,
		"subjectType": e.SubjectType,
		"subjectId":   e.SubjectID,
		"version":     e.Version,
		"modifiedAt":  e.ModifiedAt,
		"modifiedBy":  e.ModifiedBy,
	}
}

type FieldValueRecorded struct {
	FactsEventBase
	FieldID string                     `json:"fieldId"`
	Value   valueobjects.ValueEnvelope `json:"value"`
}

func NewFieldValueRecorded(params ModifyFactsParams, fieldID string, value valueobjects.ValueEnvelope) FieldValueRecorded {
	return FieldValueRecorded{FactsEventBase: newFactsEventBase(params), FieldID: fieldID, Value: value}
}

func (e FieldValueRecorded) EventType() string {
	return "FieldValueRecorded"
}

func (e FieldValueRecorded) EventData() map[string]interface{} {
	data := e.baseFactsEventData()
	data["fieldId"] = e.FieldID
	data["value"] = e.Value
	return data
}

type FieldValueCleared struct {
	FactsEventBase
	FieldID string `json:"fieldId"`
}

func NewFieldValueCleared(params ModifyFactsParams, fieldID string) FieldValueCleared {
	return FieldValueCleared{FactsEventBase: newFactsEventBase(params), FieldID: fieldID}
}

func (e FieldValueCleared) EventType() string {
	return "FieldValueCleared"
}

func (e FieldValueCleared) EventData() map[string]interface{} {
	data := e.baseFactsEventData()
	data["fieldId"] = e.FieldID
	return data
}

type OnePagerFactsArchived struct {
	FactsEventBase
	Reason string `json:"reason"`
}

func NewOnePagerFactsArchived(params ModifyFactsParams, reason string) OnePagerFactsArchived {
	return OnePagerFactsArchived{FactsEventBase: newFactsEventBase(params), Reason: reason}
}

func (e OnePagerFactsArchived) EventType() string {
	return "OnePagerFactsArchived"
}

func (e OnePagerFactsArchived) EventData() map[string]interface{} {
	data := e.baseFactsEventData()
	data["reason"] = e.Reason
	return data
}
