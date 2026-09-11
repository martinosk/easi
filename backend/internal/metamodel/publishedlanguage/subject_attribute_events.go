package publishedlanguage

import "time"

type SubjectAttributeOptionData struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

type SubjectAttributeEventBase struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	SubjectType string    `json:"subjectType"`
	Version     int       `json:"version"`
	AttributeID string    `json:"attributeId"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	ModifiedBy  string    `json:"modifiedBy"`
}

type SubjectAttributeDefinedPayload struct {
	SubjectAttributeEventBase
	Name          string                       `json:"name"`
	AttributeType string                       `json:"attributeType"`
	HelpText      string                       `json:"helpText"`
	Options       []SubjectAttributeOptionData `json:"options"`
	Min           *float64                     `json:"min,omitempty"`
	Max           *float64                     `json:"max,omitempty"`
}

type SubjectAttributeRenamedPayload struct {
	SubjectAttributeEventBase
	NewName     string `json:"newName"`
	NewHelpText string `json:"newHelpText"`
}

type SubjectAttributeRetiredPayload struct {
	SubjectAttributeEventBase
}

type SubjectAttributeReactivatedPayload struct {
	SubjectAttributeEventBase
}

type SubjectAttributeOptionAddedPayload struct {
	SubjectAttributeEventBase
	OptionID string `json:"optionId"`
	Label    string `json:"label"`
}

type SubjectAttributeOptionRetiredPayload struct {
	SubjectAttributeEventBase
	OptionID string `json:"optionId"`
}

type SubjectAttributeBoundsChangedPayload struct {
	SubjectAttributeEventBase
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}
