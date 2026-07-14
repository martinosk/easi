package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type RealizationRoleAssigned struct {
	domain.BaseEvent
	ID                   string    `json:"id"`
	CapabilityID         string    `json:"capabilityId"`
	ComponentID          string    `json:"componentId"`
	RealizationID        string    `json:"realizationId"`
	Role                 string    `json:"role"`
	DisplacedComponentID string    `json:"displacedComponentId,omitempty"`
	AssignedBy           string    `json:"assignedBy"`
	OccurredOn           time.Time `json:"occurredOn"`
}

type RealizationRoleAssignedFields struct {
	ID                   string
	CapabilityID         string
	ComponentID          string
	RealizationID        string
	Role                 string
	DisplacedComponentID string
	AssignedBy           string
}

func NewRealizationRoleAssigned(f RealizationRoleAssignedFields) RealizationRoleAssigned {
	return RealizationRoleAssigned{
		BaseEvent:            domain.NewBaseEvent(f.ID),
		ID:                   f.ID,
		CapabilityID:         f.CapabilityID,
		ComponentID:          f.ComponentID,
		RealizationID:        f.RealizationID,
		Role:                 f.Role,
		DisplacedComponentID: f.DisplacedComponentID,
		AssignedBy:           f.AssignedBy,
		OccurredOn:           time.Now().UTC(),
	}
}

func (e RealizationRoleAssigned) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e RealizationRoleAssigned) EventType() string { return pl.RealizationRoleAssigned }

func (e RealizationRoleAssigned) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                   e.ID,
		"capabilityId":         e.CapabilityID,
		"componentId":          e.ComponentID,
		"realizationId":        e.RealizationID,
		"role":                 e.Role,
		"displacedComponentId": e.DisplacedComponentID,
		"assignedBy":           e.AssignedBy,
		"occurredOn":           e.OccurredOn,
	}
}
