package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type ImportSessionCreated struct {
	domain.BaseEvent
	ID                string                 `json:"id"`
	SourceFormat      string                 `json:"sourceFormat"`
	BusinessDomainID  string                 `json:"businessDomainId"`
	CapabilityEAOwner string                 `json:"capabilityEAOwner"`
	Preview           map[string]interface{} `json:"preview"`
	ParsedData        map[string]interface{} `json:"parsedData"`
	CreatedAt         time.Time              `json:"createdAt"`
}

func (e ImportSessionCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewImportSessionCreated(id, sourceFormat, businessDomainID, capabilityEAOwner string, preview, parsedData map[string]interface{}) ImportSessionCreated {
	return ImportSessionCreated{
		BaseEvent:         domain.NewBaseEvent(id),
		ID:                id,
		SourceFormat:      sourceFormat,
		BusinessDomainID:  businessDomainID,
		CapabilityEAOwner: capabilityEAOwner,
		Preview:           preview,
		ParsedData:        parsedData,
		CreatedAt:         time.Now().UTC(),
	}
}

func (e ImportSessionCreated) EventType() string {
	return "ImportSessionCreated"
}

func (e ImportSessionCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                e.ID,
		"sourceFormat":      e.SourceFormat,
		"businessDomainId":  e.BusinessDomainID,
		"capabilityEAOwner": e.CapabilityEAOwner,
		"preview":           e.Preview,
		"parsedData":        e.ParsedData,
		"createdAt":         e.CreatedAt,
	}
}
