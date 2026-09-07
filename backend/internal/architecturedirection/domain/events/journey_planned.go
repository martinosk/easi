package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type JourneyPlanned struct {
	domain.BaseEvent
	ID               string            `json:"id"`
	CapabilityID     string            `json:"capabilityId"`
	Kind             string            `json:"kind"`
	FromComponentIDs []string          `json:"fromComponentIds"`
	ToComponentID    string            `json:"toComponentId"`
	Note             string            `json:"note"`
	TargetPeriod     *TargetPeriodData `json:"targetPeriod,omitempty"`
	TargetDomainID   string            `json:"targetDomainId,omitempty"`
	TargetParentID   string            `json:"targetParentId,omitempty"`
	ResultingName    string            `json:"resultingName,omitempty"`
	TargetMaturity   *int              `json:"targetMaturity,omitempty"`
	PlannedBy        string            `json:"plannedBy"`
	OccurredOn       time.Time         `json:"occurredOn"`
}

type JourneyPlannedFields struct {
	ID               string
	CapabilityID     string
	Kind             string
	FromComponentIDs []string
	ToComponentID    string
	Note             string
	TargetPeriod     *TargetPeriodData
	TargetDomainID   string
	TargetParentID   string
	ResultingName    string
	TargetMaturity   *int
	PlannedBy        string
}

func NewJourneyPlanned(f JourneyPlannedFields) JourneyPlanned {
	return JourneyPlanned{
		BaseEvent:        domain.NewBaseEvent(f.ID),
		ID:               f.ID,
		CapabilityID:     f.CapabilityID,
		Kind:             f.Kind,
		FromComponentIDs: f.FromComponentIDs,
		ToComponentID:    f.ToComponentID,
		Note:             f.Note,
		TargetPeriod:     f.TargetPeriod,
		TargetDomainID:   f.TargetDomainID,
		TargetParentID:   f.TargetParentID,
		ResultingName:    f.ResultingName,
		TargetMaturity:   f.TargetMaturity,
		PlannedBy:        f.PlannedBy,
		OccurredOn:       time.Now().UTC(),
	}
}

func (e JourneyPlanned) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e JourneyPlanned) EventType() string { return pl.JourneyPlanned }

func (e JourneyPlanned) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"capabilityId":     e.CapabilityID,
		"kind":             e.Kind,
		"fromComponentIds": e.FromComponentIDs,
		"toComponentId":    e.ToComponentID,
		"note":             e.Note,
		"targetPeriod":     targetPeriodEventData(e.TargetPeriod),
		"targetDomainId":   e.TargetDomainID,
		"targetParentId":   e.TargetParentID,
		"resultingName":    e.ResultingName,
		"targetMaturity":   e.TargetMaturity,
		"plannedBy":        e.PlannedBy,
		"occurredOn":       e.OccurredOn,
	}
}
