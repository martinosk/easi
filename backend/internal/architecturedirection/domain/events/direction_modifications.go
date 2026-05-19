package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type DirectionNarrativeUpdated struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	Narrative  string    `json:"narrative"`
	OccurredOn time.Time `json:"occurredOn"`
}

func NewDirectionNarrativeUpdated(id, narrative string) DirectionNarrativeUpdated {
	return DirectionNarrativeUpdated{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		Narrative:  narrative,
		OccurredOn: time.Now().UTC(),
	}
}
func (e DirectionNarrativeUpdated) EventType() string { return pl.DirectionNarrativeUpdated }
func (e DirectionNarrativeUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{"id": e.ID, "narrative": e.Narrative, "occurredOn": e.OccurredOn}
}

type DirectionHorizonChanged struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	Horizon    string    `json:"horizon"`
	OccurredOn time.Time `json:"occurredOn"`
}

func NewDirectionHorizonChanged(id, horizon string) DirectionHorizonChanged {
	return DirectionHorizonChanged{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		Horizon:    horizon,
		OccurredOn: time.Now().UTC(),
	}
}
func (e DirectionHorizonChanged) EventType() string { return pl.DirectionHorizonChanged }
func (e DirectionHorizonChanged) EventData() map[string]interface{} {
	return map[string]interface{}{"id": e.ID, "horizon": e.Horizon, "occurredOn": e.OccurredOn}
}

type DirectionPlacementsChanged struct {
	domain.BaseEvent
	ID         string          `json:"id"`
	Placements []PlacementData `json:"placements"`
	OccurredOn time.Time       `json:"occurredOn"`
}

func NewDirectionPlacementsChanged(id string, placements []PlacementData) DirectionPlacementsChanged {
	return DirectionPlacementsChanged{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		Placements: placements,
		OccurredOn: time.Now().UTC(),
	}
}
func (e DirectionPlacementsChanged) EventType() string { return pl.DirectionPlacementsChanged }
func (e DirectionPlacementsChanged) EventData() map[string]interface{} {
	placements := make([]map[string]interface{}, len(e.Placements))
	for i, p := range e.Placements {
		placements[i] = map[string]interface{}{
			"targetBusinessDomainId": p.TargetBusinessDomainID,
			"resultingName":          p.ResultingName,
		}
	}
	return map[string]interface{}{"id": e.ID, "placements": placements, "occurredOn": e.OccurredOn}
}

type DirectionSourceCapabilitiesChanged struct {
	domain.BaseEvent
	ID                  string    `json:"id"`
	SourceCapabilityIDs []string  `json:"sourceCapabilityIds"`
	OccurredOn          time.Time `json:"occurredOn"`
}

func NewDirectionSourceCapabilitiesChanged(id string, sourceCapabilityIDs []string) DirectionSourceCapabilitiesChanged {
	return DirectionSourceCapabilitiesChanged{
		BaseEvent:           domain.NewBaseEvent(id),
		ID:                  id,
		SourceCapabilityIDs: sourceCapabilityIDs,
		OccurredOn:          time.Now().UTC(),
	}
}
func (e DirectionSourceCapabilitiesChanged) EventType() string {
	return pl.DirectionSourceCapabilitiesChanged
}
func (e DirectionSourceCapabilitiesChanged) EventData() map[string]interface{} {
	return map[string]interface{}{"id": e.ID, "sourceCapabilityIds": e.SourceCapabilityIDs, "occurredOn": e.OccurredOn}
}
