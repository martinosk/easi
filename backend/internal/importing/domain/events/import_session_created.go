package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

type ImportSessionCreated struct {
	domain.BaseEvent
	ID               string
	SourceFormat     string
	BusinessDomainID string
	Preview          map[string]interface{}
	ParsedData       map[string]interface{}
	CreatedAt        time.Time
}

func NewImportSessionCreated(id, sourceFormat, businessDomainID string, preview, parsedData map[string]interface{}) ImportSessionCreated {
	return ImportSessionCreated{
		BaseEvent:        domain.NewBaseEvent(id),
		ID:               id,
		SourceFormat:     sourceFormat,
		BusinessDomainID: businessDomainID,
		Preview:          preview,
		ParsedData:       parsedData,
		CreatedAt:        time.Now().UTC(),
	}
}

func (e ImportSessionCreated) EventType() string {
	return "ImportSessionCreated"
}

func (e ImportSessionCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"sourceFormat":     e.SourceFormat,
		"businessDomainId": e.BusinessDomainID,
		"preview":          e.Preview,
		"parsedData":       e.ParsedData,
		"createdAt":        e.CreatedAt,
	}
}
