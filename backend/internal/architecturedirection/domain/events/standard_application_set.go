package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type StandardApplicationSet struct {
	domain.BaseEvent
	ID                     string    `json:"id"`
	EnterpriseCapabilityID string    `json:"enterpriseCapabilityId"`
	ApplicationID          string    `json:"applicationId"`
	PreviousApplicationID  string    `json:"previousApplicationId,omitempty"`
	Narrative              string    `json:"narrative"`
	OccurredOn             time.Time `json:"occurredOn"`
}

type StandardApplicationSetFields struct {
	ID                     string
	EnterpriseCapabilityID string
	ApplicationID          string
	PreviousApplicationID  string
	Narrative              string
}

func NewStandardApplicationSet(f StandardApplicationSetFields) StandardApplicationSet {
	return StandardApplicationSet{
		BaseEvent:              domain.NewBaseEvent(f.ID),
		ID:                     f.ID,
		EnterpriseCapabilityID: f.EnterpriseCapabilityID,
		ApplicationID:          f.ApplicationID,
		PreviousApplicationID:  f.PreviousApplicationID,
		Narrative:              f.Narrative,
		OccurredOn:             time.Now().UTC(),
	}
}

func (e StandardApplicationSet) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e StandardApplicationSet) EventType() string { return pl.StandardApplicationSet }

func (e StandardApplicationSet) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                     e.ID,
		"enterpriseCapabilityId": e.EnterpriseCapabilityID,
		"applicationId":          e.ApplicationID,
		"previousApplicationId":  e.PreviousApplicationID,
		"narrative":              e.Narrative,
		"occurredOn":             e.OccurredOn,
	}
}
