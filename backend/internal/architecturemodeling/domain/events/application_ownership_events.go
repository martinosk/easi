package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationOwnerNominated struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	OwnerKind   string    `json:"ownerKind"`
	OwnerID     string    `json:"ownerId"`
	NominatedAt time.Time `json:"nominatedAt"`
}

func NewApplicationOwnerNominated(componentID, ownerKind, ownerID string) ApplicationOwnerNominated {
	return ApplicationOwnerNominated{
		BaseEvent:   domain.NewBaseEvent(componentID),
		ComponentID: componentID,
		OwnerKind:   ownerKind,
		OwnerID:     ownerID,
		NominatedAt: time.Now().UTC(),
	}
}

func (e ApplicationOwnerNominated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func (e ApplicationOwnerNominated) EventType() string {
	return "ApplicationOwnerNominated"
}

func (e ApplicationOwnerNominated) EventData() map[string]any {
	return map[string]any{
		"componentId": e.ComponentID,
		"ownerKind":   e.OwnerKind,
		"ownerId":     e.OwnerID,
		"nominatedAt": e.NominatedAt,
	}
}

type ApplicationOwnershipConfirmed struct {
	domain.BaseEvent
	ComponentID    string    `json:"componentId"`
	OwnerKind      string    `json:"ownerKind"`
	OwnerID        string    `json:"ownerId"`
	OwnershipState string    `json:"ownershipState"`
	ConfirmedAt    time.Time `json:"confirmedAt"`
}

func NewApplicationOwnershipConfirmed(componentID, ownerKind, ownerID, ownershipState string) ApplicationOwnershipConfirmed {
	return ApplicationOwnershipConfirmed{
		BaseEvent:      domain.NewBaseEvent(componentID),
		ComponentID:    componentID,
		OwnerKind:      ownerKind,
		OwnerID:        ownerID,
		OwnershipState: ownershipState,
		ConfirmedAt:    time.Now().UTC(),
	}
}

func (e ApplicationOwnershipConfirmed) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func (e ApplicationOwnershipConfirmed) EventType() string {
	return "ApplicationOwnershipConfirmed"
}

func (e ApplicationOwnershipConfirmed) EventData() map[string]any {
	return map[string]any{
		"componentId":    e.ComponentID,
		"ownerKind":      e.OwnerKind,
		"ownerId":        e.OwnerID,
		"ownershipState": e.OwnershipState,
		"confirmedAt":    e.ConfirmedAt,
	}
}

type ApplicationOwnerAssigned struct {
	domain.BaseEvent
	ComponentID    string    `json:"componentId"`
	OwnerKind      string    `json:"ownerKind"`
	OwnerID        string    `json:"ownerId"`
	OwnershipState string    `json:"ownershipState"`
	AssignedAt     time.Time `json:"assignedAt"`
}

func NewApplicationOwnerAssigned(componentID, ownerKind, ownerID, ownershipState string) ApplicationOwnerAssigned {
	return ApplicationOwnerAssigned{
		BaseEvent:      domain.NewBaseEvent(componentID),
		ComponentID:    componentID,
		OwnerKind:      ownerKind,
		OwnerID:        ownerID,
		OwnershipState: ownershipState,
		AssignedAt:     time.Now().UTC(),
	}
}

func (e ApplicationOwnerAssigned) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func (e ApplicationOwnerAssigned) EventType() string {
	return "ApplicationOwnerAssigned"
}

func (e ApplicationOwnerAssigned) EventData() map[string]any {
	return map[string]any{
		"componentId":    e.ComponentID,
		"ownerKind":      e.OwnerKind,
		"ownerId":        e.OwnerID,
		"ownershipState": e.OwnershipState,
		"assignedAt":     e.AssignedAt,
	}
}

type ApplicationOwnershipCleared struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	ClearedAt   time.Time `json:"clearedAt"`
}

func NewApplicationOwnershipCleared(componentID string) ApplicationOwnershipCleared {
	return ApplicationOwnershipCleared{
		BaseEvent:   domain.NewBaseEvent(componentID),
		ComponentID: componentID,
		ClearedAt:   time.Now().UTC(),
	}
}

func (e ApplicationOwnershipCleared) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func (e ApplicationOwnershipCleared) EventType() string {
	return "ApplicationOwnershipCleared"
}

func (e ApplicationOwnershipCleared) EventData() map[string]any {
	return map[string]any{
		"componentId": e.ComponentID,
		"clearedAt":   e.ClearedAt,
	}
}
