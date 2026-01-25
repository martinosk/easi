package events

import (
	"time"

	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ComponentOriginsCreated struct {
	domain.BaseEvent
	AggregateIDValue string    `json:"aggregateId"`
	ComponentID      string    `json:"componentId"`
	CreatedAt        time.Time `json:"createdAt"`
}

func (e ComponentOriginsCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.AggregateIDValue
}

func NewComponentOriginsCreatedEvent(aggregateID string, componentID valueobjects.ComponentID, createdAt time.Time) ComponentOriginsCreated {
	return ComponentOriginsCreated{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		CreatedAt:   createdAt,
	}
}

func (e ComponentOriginsCreated) EventType() string {
	return "ComponentOriginsCreated"
}

func (e ComponentOriginsCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"aggregateId": e.AggregateID(),
		"componentId": e.ComponentID,
		"createdAt":   e.CreatedAt,
	}
}

type AcquiredViaRelationshipSet struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	EntityID    string    `json:"entityId"`
	Notes       string    `json:"notes"`
	LinkedAt    time.Time `json:"linkedAt"`
}

func (e AcquiredViaRelationshipSet) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewAcquiredViaRelationshipSetEvent(aggregateID string, componentID valueobjects.ComponentID, entityID valueobjects.AcquiredEntityID, notes valueobjects.Notes, linkedAt time.Time) AcquiredViaRelationshipSet {
	return AcquiredViaRelationshipSet{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		EntityID:    entityID.String(),
		Notes:       notes.String(),
		LinkedAt:    linkedAt,
	}
}

func (e AcquiredViaRelationshipSet) EventType() string {
	return "AcquiredViaRelationshipSet"
}

func (e AcquiredViaRelationshipSet) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"entityId":    e.EntityID,
		"notes":       e.Notes,
		"linkedAt":    e.LinkedAt,
	}
}

type AcquiredViaRelationshipReplaced struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	OldEntityID string    `json:"oldEntityId"`
	NewEntityID string    `json:"newEntityId"`
	Notes       string    `json:"notes"`
	LinkedAt    time.Time `json:"linkedAt"`
}

func (e AcquiredViaRelationshipReplaced) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewAcquiredViaRelationshipReplacedEvent(aggregateID string, componentID valueobjects.ComponentID, oldEntityID, newEntityID valueobjects.AcquiredEntityID, notes valueobjects.Notes, linkedAt time.Time) AcquiredViaRelationshipReplaced {
	return AcquiredViaRelationshipReplaced{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		OldEntityID: oldEntityID.String(),
		NewEntityID: newEntityID.String(),
		Notes:       notes.String(),
		LinkedAt:    linkedAt,
	}
}

func (e AcquiredViaRelationshipReplaced) EventType() string {
	return "AcquiredViaRelationshipReplaced"
}

func (e AcquiredViaRelationshipReplaced) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"oldEntityId": e.OldEntityID,
		"newEntityId": e.NewEntityID,
		"notes":       e.Notes,
		"linkedAt":    e.LinkedAt,
	}
}

type AcquiredViaNotesUpdated struct {
	domain.BaseEvent
	ComponentID string `json:"componentId"`
	EntityID    string `json:"entityId"`
	OldNotes    string `json:"oldNotes"`
	NewNotes    string `json:"newNotes"`
}

func (e AcquiredViaNotesUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewAcquiredViaNotesUpdatedEvent(aggregateID string, componentID valueobjects.ComponentID, entityID valueobjects.AcquiredEntityID, oldNotes, newNotes valueobjects.Notes) AcquiredViaNotesUpdated {
	return AcquiredViaNotesUpdated{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		EntityID:    entityID.String(),
		OldNotes:    oldNotes.String(),
		NewNotes:    newNotes.String(),
	}
}

func (e AcquiredViaNotesUpdated) EventType() string {
	return "AcquiredViaNotesUpdated"
}

func (e AcquiredViaNotesUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"entityId":    e.EntityID,
		"oldNotes":    e.OldNotes,
		"newNotes":    e.NewNotes,
	}
}

type AcquiredViaRelationshipCleared struct {
	domain.BaseEvent
	ComponentID string `json:"componentId"`
	EntityID    string `json:"entityId"`
}

func (e AcquiredViaRelationshipCleared) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewAcquiredViaRelationshipClearedEvent(aggregateID string, componentID valueobjects.ComponentID, entityID valueobjects.AcquiredEntityID) AcquiredViaRelationshipCleared {
	return AcquiredViaRelationshipCleared{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		EntityID:    entityID.String(),
	}
}

func (e AcquiredViaRelationshipCleared) EventType() string {
	return "AcquiredViaRelationshipCleared"
}

func (e AcquiredViaRelationshipCleared) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"entityId":    e.EntityID,
	}
}

type PurchasedFromRelationshipSet struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	VendorID    string    `json:"vendorId"`
	Notes       string    `json:"notes"`
	LinkedAt    time.Time `json:"linkedAt"`
}

func (e PurchasedFromRelationshipSet) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewPurchasedFromRelationshipSetEvent(aggregateID string, componentID valueobjects.ComponentID, vendorID valueobjects.VendorID, notes valueobjects.Notes, linkedAt time.Time) PurchasedFromRelationshipSet {
	return PurchasedFromRelationshipSet{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		VendorID:    vendorID.String(),
		Notes:       notes.String(),
		LinkedAt:    linkedAt,
	}
}

func (e PurchasedFromRelationshipSet) EventType() string {
	return "PurchasedFromRelationshipSet"
}

func (e PurchasedFromRelationshipSet) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"vendorId":    e.VendorID,
		"notes":       e.Notes,
		"linkedAt":    e.LinkedAt,
	}
}

type PurchasedFromRelationshipReplaced struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	OldVendorID string    `json:"oldVendorId"`
	NewVendorID string    `json:"newVendorId"`
	Notes       string    `json:"notes"`
	LinkedAt    time.Time `json:"linkedAt"`
}

func (e PurchasedFromRelationshipReplaced) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewPurchasedFromRelationshipReplacedEvent(aggregateID string, componentID valueobjects.ComponentID, oldVendorID, newVendorID valueobjects.VendorID, notes valueobjects.Notes, linkedAt time.Time) PurchasedFromRelationshipReplaced {
	return PurchasedFromRelationshipReplaced{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		OldVendorID: oldVendorID.String(),
		NewVendorID: newVendorID.String(),
		Notes:       notes.String(),
		LinkedAt:    linkedAt,
	}
}

func (e PurchasedFromRelationshipReplaced) EventType() string {
	return "PurchasedFromRelationshipReplaced"
}

func (e PurchasedFromRelationshipReplaced) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"oldVendorId": e.OldVendorID,
		"newVendorId": e.NewVendorID,
		"notes":       e.Notes,
		"linkedAt":    e.LinkedAt,
	}
}

type PurchasedFromNotesUpdated struct {
	domain.BaseEvent
	ComponentID string `json:"componentId"`
	VendorID    string `json:"vendorId"`
	OldNotes    string `json:"oldNotes"`
	NewNotes    string `json:"newNotes"`
}

func (e PurchasedFromNotesUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewPurchasedFromNotesUpdatedEvent(aggregateID string, componentID valueobjects.ComponentID, vendorID valueobjects.VendorID, oldNotes, newNotes valueobjects.Notes) PurchasedFromNotesUpdated {
	return PurchasedFromNotesUpdated{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		VendorID:    vendorID.String(),
		OldNotes:    oldNotes.String(),
		NewNotes:    newNotes.String(),
	}
}

func (e PurchasedFromNotesUpdated) EventType() string {
	return "PurchasedFromNotesUpdated"
}

func (e PurchasedFromNotesUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"vendorId":    e.VendorID,
		"oldNotes":    e.OldNotes,
		"newNotes":    e.NewNotes,
	}
}

type PurchasedFromRelationshipCleared struct {
	domain.BaseEvent
	ComponentID string `json:"componentId"`
	VendorID    string `json:"vendorId"`
}

func (e PurchasedFromRelationshipCleared) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewPurchasedFromRelationshipClearedEvent(aggregateID string, componentID valueobjects.ComponentID, vendorID valueobjects.VendorID) PurchasedFromRelationshipCleared {
	return PurchasedFromRelationshipCleared{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		VendorID:    vendorID.String(),
	}
}

func (e PurchasedFromRelationshipCleared) EventType() string {
	return "PurchasedFromRelationshipCleared"
}

func (e PurchasedFromRelationshipCleared) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"vendorId":    e.VendorID,
	}
}

type BuiltByRelationshipSet struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	TeamID      string    `json:"teamId"`
	Notes       string    `json:"notes"`
	LinkedAt    time.Time `json:"linkedAt"`
}

func (e BuiltByRelationshipSet) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewBuiltByRelationshipSetEvent(aggregateID string, componentID valueobjects.ComponentID, teamID valueobjects.InternalTeamID, notes valueobjects.Notes, linkedAt time.Time) BuiltByRelationshipSet {
	return BuiltByRelationshipSet{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		TeamID:      teamID.String(),
		Notes:       notes.String(),
		LinkedAt:    linkedAt,
	}
}

func (e BuiltByRelationshipSet) EventType() string {
	return "BuiltByRelationshipSet"
}

func (e BuiltByRelationshipSet) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"teamId":      e.TeamID,
		"notes":       e.Notes,
		"linkedAt":    e.LinkedAt,
	}
}

type BuiltByRelationshipReplaced struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	OldTeamID   string    `json:"oldTeamId"`
	NewTeamID   string    `json:"newTeamId"`
	Notes       string    `json:"notes"`
	LinkedAt    time.Time `json:"linkedAt"`
}

func (e BuiltByRelationshipReplaced) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewBuiltByRelationshipReplacedEvent(aggregateID string, componentID valueobjects.ComponentID, oldTeamID, newTeamID valueobjects.InternalTeamID, notes valueobjects.Notes, linkedAt time.Time) BuiltByRelationshipReplaced {
	return BuiltByRelationshipReplaced{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		OldTeamID:   oldTeamID.String(),
		NewTeamID:   newTeamID.String(),
		Notes:       notes.String(),
		LinkedAt:    linkedAt,
	}
}

func (e BuiltByRelationshipReplaced) EventType() string {
	return "BuiltByRelationshipReplaced"
}

func (e BuiltByRelationshipReplaced) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"oldTeamId":   e.OldTeamID,
		"newTeamId":   e.NewTeamID,
		"notes":       e.Notes,
		"linkedAt":    e.LinkedAt,
	}
}

type BuiltByNotesUpdated struct {
	domain.BaseEvent
	ComponentID string `json:"componentId"`
	TeamID      string `json:"teamId"`
	OldNotes    string `json:"oldNotes"`
	NewNotes    string `json:"newNotes"`
}

func (e BuiltByNotesUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewBuiltByNotesUpdatedEvent(aggregateID string, componentID valueobjects.ComponentID, teamID valueobjects.InternalTeamID, oldNotes, newNotes valueobjects.Notes) BuiltByNotesUpdated {
	return BuiltByNotesUpdated{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		TeamID:      teamID.String(),
		OldNotes:    oldNotes.String(),
		NewNotes:    newNotes.String(),
	}
}

func (e BuiltByNotesUpdated) EventType() string {
	return "BuiltByNotesUpdated"
}

func (e BuiltByNotesUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"teamId":      e.TeamID,
		"oldNotes":    e.OldNotes,
		"newNotes":    e.NewNotes,
	}
}

type BuiltByRelationshipCleared struct {
	domain.BaseEvent
	ComponentID string `json:"componentId"`
	TeamID      string `json:"teamId"`
}

func (e BuiltByRelationshipCleared) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewBuiltByRelationshipClearedEvent(aggregateID string, componentID valueobjects.ComponentID, teamID valueobjects.InternalTeamID) BuiltByRelationshipCleared {
	return BuiltByRelationshipCleared{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		TeamID:      teamID.String(),
	}
}

func (e BuiltByRelationshipCleared) EventType() string {
	return "BuiltByRelationshipCleared"
}

func (e BuiltByRelationshipCleared) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"teamId":      e.TeamID,
	}
}

type ComponentOriginsDeleted struct {
	domain.BaseEvent
	ComponentID string    `json:"componentId"`
	DeletedAt   time.Time `json:"deletedAt"`
}

func (e ComponentOriginsDeleted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ComponentID
}

func NewComponentOriginsDeletedEvent(aggregateID string, componentID valueobjects.ComponentID, deletedAt time.Time) ComponentOriginsDeleted {
	return ComponentOriginsDeleted{
		BaseEvent:   domain.NewBaseEvent(aggregateID),
		ComponentID: componentID.String(),
		DeletedAt:   deletedAt,
	}
}

func (e ComponentOriginsDeleted) EventType() string {
	return "ComponentOriginsDeleted"
}

func (e ComponentOriginsDeleted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"componentId": e.ComponentID,
		"deletedAt":   e.DeletedAt,
	}
}
