package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type PlacementData struct {
	TargetBusinessDomainID string `json:"targetBusinessDomainId"`
	ResultingName          string `json:"resultingName,omitempty"`
}

type DirectionDrafted struct {
	domain.BaseEvent
	ID                     string          `json:"id"`
	EnterpriseCapabilityID string          `json:"enterpriseCapabilityId"`
	Type                   string          `json:"type"`
	SourceCapabilityIDs    []string        `json:"sourceCapabilityIds"`
	Placements             []PlacementData `json:"placements"`
	Horizon                string          `json:"horizon"`
	Narrative              string          `json:"narrative,omitempty"`
	CreatedAt              time.Time       `json:"createdAt"`
}

type DirectionDraftedFields struct {
	ID                     string
	EnterpriseCapabilityID string
	Type                   string
	SourceCapabilityIDs    []string
	Placements             []PlacementData
	Horizon                string
	Narrative              string
}

func (e DirectionDrafted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewDirectionDraftedFromFields(f DirectionDraftedFields) DirectionDrafted {
	return DirectionDrafted{
		BaseEvent:              domain.NewBaseEvent(f.ID),
		ID:                     f.ID,
		EnterpriseCapabilityID: f.EnterpriseCapabilityID,
		Type:                   f.Type,
		SourceCapabilityIDs:    f.SourceCapabilityIDs,
		Placements:             f.Placements,
		Horizon:                f.Horizon,
		Narrative:              f.Narrative,
		CreatedAt:              time.Now().UTC(),
	}
}

func (e DirectionDrafted) EventType() string { return pl.DirectionDrafted }

func (e DirectionDrafted) EventData() map[string]interface{} {
	placements := make([]map[string]interface{}, len(e.Placements))
	for i, p := range e.Placements {
		placements[i] = map[string]interface{}{
			"targetBusinessDomainId": p.TargetBusinessDomainID,
			"resultingName":          p.ResultingName,
		}
	}
	return map[string]interface{}{
		"id":                     e.ID,
		"enterpriseCapabilityId": e.EnterpriseCapabilityID,
		"type":                   e.Type,
		"sourceCapabilityIds":    e.SourceCapabilityIDs,
		"placements":             placements,
		"horizon":                e.Horizon,
		"narrative":              e.Narrative,
		"createdAt":              e.CreatedAt,
	}
}
