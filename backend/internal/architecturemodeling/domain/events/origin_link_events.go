package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type OriginLinkCreated struct {
	OriginLinkBase
	AggregateIDValue string    `json:"aggregateId"`
	CreatedAt        time.Time `json:"createdAt"`
}

func (e OriginLinkCreated) AggregateID() string {
	if baseID := e.OriginLinkBase.aggregateID(); baseID != "" {
		return baseID
	}
	return e.AggregateIDValue
}

func NewOriginLinkCreatedEvent(base OriginLinkBase, createdAt time.Time) OriginLinkCreated {
	return OriginLinkCreated{
		OriginLinkBase: base,
		CreatedAt:      createdAt,
	}
}

func (e OriginLinkCreated) EventType() string { return "OriginLinkCreated" }

func (e OriginLinkCreated) EventData() map[string]interface{} {
	data := e.OriginLinkBase.eventData()
	data["aggregateId"] = e.AggregateID()
	data["createdAt"] = e.CreatedAt
	return data
}

type OriginLinkBase struct {
	domain.BaseEvent
	ComponentID string `json:"componentId"`
	OriginType  string `json:"originType"`
}

func NewOriginLinkBase(aggregateID, componentID, originType string) OriginLinkBase {
	return OriginLinkBase{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID,
		OriginType:  originType,
	}
}

func (b OriginLinkBase) aggregateID() string {
	if baseID := b.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return b.ComponentID
}

func (b OriginLinkBase) eventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": b.ComponentID,
		"originType":  b.OriginType,
	}
}

type OriginLinkSet struct {
	OriginLinkBase
	EntityID string    `json:"entityId"`
	Notes    string    `json:"notes"`
	LinkedAt time.Time `json:"linkedAt"`
}

func (e OriginLinkSet) AggregateID() string { return e.OriginLinkBase.aggregateID() }

func NewOriginLinkSetEvent(base OriginLinkBase, entityID, notes string, linkedAt time.Time) OriginLinkSet {
	return OriginLinkSet{
		OriginLinkBase: base,
		EntityID:       entityID,
		Notes:          notes,
		LinkedAt:       linkedAt,
	}
}

func (e OriginLinkSet) EventType() string { return "OriginLinkSet" }

func (e OriginLinkSet) EventData() map[string]interface{} {
	data := e.OriginLinkBase.eventData()
	data["entityId"] = e.EntityID
	data["notes"] = e.Notes
	data["linkedAt"] = e.LinkedAt
	return data
}

type OriginLinkReplacement struct {
	OldEntityID string    `json:"oldEntityId"`
	NewEntityID string    `json:"newEntityId"`
	Notes       string    `json:"notes"`
	LinkedAt    time.Time `json:"linkedAt"`
}

type OriginLinkReplaced struct {
	OriginLinkBase
	OriginLinkReplacement
}

func (e OriginLinkReplaced) AggregateID() string { return e.OriginLinkBase.aggregateID() }

func NewOriginLinkReplacedEvent(base OriginLinkBase, replacement OriginLinkReplacement) OriginLinkReplaced {
	return OriginLinkReplaced{
		OriginLinkBase:        base,
		OriginLinkReplacement: replacement,
	}
}

func (e OriginLinkReplaced) EventType() string { return "OriginLinkReplaced" }

func (e OriginLinkReplaced) EventData() map[string]interface{} {
	data := e.OriginLinkBase.eventData()
	data["oldEntityId"] = e.OldEntityID
	data["newEntityId"] = e.NewEntityID
	data["notes"] = e.Notes
	data["linkedAt"] = e.LinkedAt
	return data
}

type OriginLinkNotesUpdated struct {
	OriginLinkBase
	EntityID string `json:"entityId"`
	OldNotes string `json:"oldNotes"`
	NewNotes string `json:"newNotes"`
}

func (e OriginLinkNotesUpdated) AggregateID() string { return e.OriginLinkBase.aggregateID() }

func NewOriginLinkNotesUpdatedEvent(base OriginLinkBase, entityID, oldNotes, newNotes string) OriginLinkNotesUpdated {
	return OriginLinkNotesUpdated{
		OriginLinkBase: base,
		EntityID:       entityID,
		OldNotes:       oldNotes,
		NewNotes:       newNotes,
	}
}

func (e OriginLinkNotesUpdated) EventType() string { return "OriginLinkNotesUpdated" }

func (e OriginLinkNotesUpdated) EventData() map[string]interface{} {
	data := e.OriginLinkBase.eventData()
	data["entityId"] = e.EntityID
	data["oldNotes"] = e.OldNotes
	data["newNotes"] = e.NewNotes
	return data
}

type OriginLinkCleared struct {
	OriginLinkBase
	EntityID string `json:"entityId"`
}

func (e OriginLinkCleared) AggregateID() string { return e.OriginLinkBase.aggregateID() }

func NewOriginLinkClearedEvent(base OriginLinkBase, entityID string) OriginLinkCleared {
	return OriginLinkCleared{
		OriginLinkBase: base,
		EntityID:       entityID,
	}
}

func (e OriginLinkCleared) EventType() string { return "OriginLinkCleared" }

func (e OriginLinkCleared) EventData() map[string]interface{} {
	data := e.OriginLinkBase.eventData()
	data["entityId"] = e.EntityID
	return data
}

type OriginLinkDeleted struct {
	OriginLinkBase
	DeletedAt time.Time `json:"deletedAt"`
}

func (e OriginLinkDeleted) AggregateID() string { return e.OriginLinkBase.aggregateID() }

func NewOriginLinkDeletedEvent(base OriginLinkBase, deletedAt time.Time) OriginLinkDeleted {
	return OriginLinkDeleted{
		OriginLinkBase: base,
		DeletedAt:      deletedAt,
	}
}

func (e OriginLinkDeleted) EventType() string { return "OriginLinkDeleted" }

func (e OriginLinkDeleted) EventData() map[string]interface{} {
	data := e.OriginLinkBase.eventData()
	data["deletedAt"] = e.DeletedAt
	return data
}
