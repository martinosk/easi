package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

const (
	TypeSubjectAttributeSchemaCreated = "SubjectAttributeSchemaCreated"
	TypeSubjectAttributeDefined       = "SubjectAttributeDefined"
	TypeSubjectAttributeRenamed       = "SubjectAttributeRenamed"
	TypeSubjectAttributeRetired       = "SubjectAttributeRetired"
	TypeSubjectAttributeReactivated   = "SubjectAttributeReactivated"
	TypeSubjectAttributeOptionAdded   = "SubjectAttributeOptionAdded"
	TypeSubjectAttributeOptionRetired = "SubjectAttributeOptionRetired"
	TypeSubjectAttributeBoundsChanged = "SubjectAttributeBoundsChanged"
)

func SubjectAttributeSchemaEventTypes() []string {
	return []string{
		TypeSubjectAttributeSchemaCreated,
		TypeSubjectAttributeDefined,
		TypeSubjectAttributeRenamed,
		TypeSubjectAttributeRetired,
		TypeSubjectAttributeReactivated,
		TypeSubjectAttributeOptionAdded,
		TypeSubjectAttributeOptionRetired,
		TypeSubjectAttributeBoundsChanged,
	}
}

type AttributeOptionData struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

type SubjectAttributeSchemaCreated struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	SubjectType string    `json:"subjectType"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
}

type CreateSchemaParams struct {
	ID          string
	TenantID    string
	SubjectType string
	CreatedBy   string
}

func NewSubjectAttributeSchemaCreated(params CreateSchemaParams) SubjectAttributeSchemaCreated {
	return SubjectAttributeSchemaCreated{
		BaseEvent:   domain.NewBaseEvent(params.ID),
		ID:          params.ID,
		TenantID:    params.TenantID,
		SubjectType: params.SubjectType,
		CreatedAt:   time.Now().UTC(),
		CreatedBy:   params.CreatedBy,
	}
}

func (e SubjectAttributeSchemaCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e SubjectAttributeSchemaCreated) EventType() string {
	return TypeSubjectAttributeSchemaCreated
}

func (e SubjectAttributeSchemaCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"tenantId":    e.TenantID,
		"subjectType": e.SubjectType,
		"createdAt":   e.CreatedAt,
		"createdBy":   e.CreatedBy,
	}
}

type SchemaEventBase struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	SubjectType string    `json:"subjectType"`
	Version     int       `json:"version"`
	AttributeID string    `json:"attributeId"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	ModifiedBy  string    `json:"modifiedBy"`
}

func (e SchemaEventBase) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

type ModifySchemaParams struct {
	SchemaID    string
	TenantID    string
	SubjectType string
	Version     int
	AttributeID string
	ModifiedBy  string
}

func newSchemaEventBase(params ModifySchemaParams) SchemaEventBase {
	return SchemaEventBase{
		BaseEvent:   domain.NewBaseEvent(params.SchemaID),
		ID:          params.SchemaID,
		TenantID:    params.TenantID,
		SubjectType: params.SubjectType,
		Version:     params.Version,
		AttributeID: params.AttributeID,
		ModifiedAt:  time.Now().UTC(),
		ModifiedBy:  params.ModifiedBy,
	}
}

func (e SchemaEventBase) baseEventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"tenantId":    e.TenantID,
		"subjectType": e.SubjectType,
		"version":     e.Version,
		"attributeId": e.AttributeID,
		"modifiedAt":  e.ModifiedAt,
		"modifiedBy":  e.ModifiedBy,
	}
}

type SubjectAttributeDefined struct {
	SchemaEventBase
	Name          string                `json:"name"`
	AttributeType string                `json:"attributeType"`
	HelpText      string                `json:"helpText"`
	Options       []AttributeOptionData `json:"options"`
	Min           *float64              `json:"min,omitempty"`
	Max           *float64              `json:"max,omitempty"`
}

type AttributeDefinitionData struct {
	Name          string
	AttributeType string
	HelpText      string
	Options       []AttributeOptionData
	Min           *float64
	Max           *float64
}

func NewSubjectAttributeDefined(params ModifySchemaParams, definition AttributeDefinitionData) SubjectAttributeDefined {
	return SubjectAttributeDefined{
		SchemaEventBase: newSchemaEventBase(params),
		Name:            definition.Name,
		AttributeType:   definition.AttributeType,
		HelpText:        definition.HelpText,
		Options:         definition.Options,
		Min:             definition.Min,
		Max:             definition.Max,
	}
}

func (e SubjectAttributeDefined) EventType() string {
	return TypeSubjectAttributeDefined
}

func (e SubjectAttributeDefined) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["name"] = e.Name
	data["attributeType"] = e.AttributeType
	data["helpText"] = e.HelpText
	data["options"] = e.Options
	data["min"] = e.Min
	data["max"] = e.Max
	return data
}

type SubjectAttributeRenamed struct {
	SchemaEventBase
	NewName     string `json:"newName"`
	NewHelpText string `json:"newHelpText"`
}

func NewSubjectAttributeRenamed(params ModifySchemaParams, newName, newHelpText string) SubjectAttributeRenamed {
	return SubjectAttributeRenamed{SchemaEventBase: newSchemaEventBase(params), NewName: newName, NewHelpText: newHelpText}
}

func (e SubjectAttributeRenamed) EventType() string {
	return TypeSubjectAttributeRenamed
}

func (e SubjectAttributeRenamed) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["newName"] = e.NewName
	data["newHelpText"] = e.NewHelpText
	return data
}

type SubjectAttributeRetired struct {
	SchemaEventBase
}

func NewSubjectAttributeRetired(params ModifySchemaParams) SubjectAttributeRetired {
	return SubjectAttributeRetired{SchemaEventBase: newSchemaEventBase(params)}
}

func (e SubjectAttributeRetired) EventType() string {
	return TypeSubjectAttributeRetired
}

func (e SubjectAttributeRetired) EventData() map[string]interface{} {
	return e.baseEventData()
}

type SubjectAttributeReactivated struct {
	SchemaEventBase
}

func NewSubjectAttributeReactivated(params ModifySchemaParams) SubjectAttributeReactivated {
	return SubjectAttributeReactivated{SchemaEventBase: newSchemaEventBase(params)}
}

func (e SubjectAttributeReactivated) EventType() string {
	return TypeSubjectAttributeReactivated
}

func (e SubjectAttributeReactivated) EventData() map[string]interface{} {
	return e.baseEventData()
}

type SubjectAttributeOptionAdded struct {
	SchemaEventBase
	OptionID string `json:"optionId"`
	Label    string `json:"label"`
}

func NewSubjectAttributeOptionAdded(params ModifySchemaParams, optionID, label string) SubjectAttributeOptionAdded {
	return SubjectAttributeOptionAdded{SchemaEventBase: newSchemaEventBase(params), OptionID: optionID, Label: label}
}

func (e SubjectAttributeOptionAdded) EventType() string {
	return TypeSubjectAttributeOptionAdded
}

func (e SubjectAttributeOptionAdded) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["optionId"] = e.OptionID
	data["label"] = e.Label
	return data
}

type SubjectAttributeOptionRetired struct {
	SchemaEventBase
	OptionID string `json:"optionId"`
}

func NewSubjectAttributeOptionRetired(params ModifySchemaParams, optionID string) SubjectAttributeOptionRetired {
	return SubjectAttributeOptionRetired{SchemaEventBase: newSchemaEventBase(params), OptionID: optionID}
}

func (e SubjectAttributeOptionRetired) EventType() string {
	return TypeSubjectAttributeOptionRetired
}

func (e SubjectAttributeOptionRetired) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["optionId"] = e.OptionID
	return data
}

type SubjectAttributeBoundsChanged struct {
	SchemaEventBase
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

func NewSubjectAttributeBoundsChanged(params ModifySchemaParams, min, max *float64) SubjectAttributeBoundsChanged {
	return SubjectAttributeBoundsChanged{SchemaEventBase: newSchemaEventBase(params), Min: min, Max: max}
}

func (e SubjectAttributeBoundsChanged) EventType() string {
	return TypeSubjectAttributeBoundsChanged
}

func (e SubjectAttributeBoundsChanged) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["min"] = e.Min
	data["max"] = e.Max
	return data
}
