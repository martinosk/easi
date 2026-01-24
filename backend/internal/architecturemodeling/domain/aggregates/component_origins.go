package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrNoAcquiredViaRelationship   = errors.New("no acquired-via relationship exists")
	ErrNoPurchasedFromRelationship = errors.New("no purchased-from relationship exists")
	ErrNoBuiltByRelationship       = errors.New("no built-by relationship exists")
)

type ComponentOrigins struct {
	domain.AggregateRoot
	componentID   valueobjects.ComponentID
	acquiredVia   valueobjects.OriginLink
	purchasedFrom valueobjects.OriginLink
	builtBy       valueobjects.OriginLink
	createdAt     time.Time
	isDeleted     bool
}

func NewComponentOrigins(componentID valueobjects.ComponentID) (*ComponentOrigins, error) {
	aggregateID := "component-origins:" + componentID.String()
	co := &ComponentOrigins{
		AggregateRoot: domain.NewAggregateRootWithID(aggregateID),
	}
	event := events.NewComponentOriginsCreatedEvent(aggregateID, componentID, time.Now())
	co.apply(event)
	co.RaiseEvent(event)
	return co, nil
}

func LoadComponentOriginsFromHistory(events []domain.DomainEvent) (*ComponentOrigins, error) {
	co := &ComponentOrigins{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	co.LoadFromHistory(events, func(event domain.DomainEvent) {
		co.apply(event)
	})

	return co, nil
}

func (co *ComponentOrigins) ComponentID() valueobjects.ComponentID {
	return co.componentID
}

func (co *ComponentOrigins) AcquiredVia() valueobjects.OriginLink {
	return co.acquiredVia
}

func (co *ComponentOrigins) PurchasedFrom() valueobjects.OriginLink {
	return co.purchasedFrom
}

func (co *ComponentOrigins) BuiltBy() valueobjects.OriginLink {
	return co.builtBy
}

func (co *ComponentOrigins) CreatedAt() time.Time {
	return co.createdAt
}

func (co *ComponentOrigins) IsDeleted() bool {
	return co.isDeleted
}

func (co *ComponentOrigins) SetAcquiredVia(entityID valueobjects.AcquiredEntityID, notes valueobjects.Notes) error {
	linkedAt := time.Now().UTC()

	if co.createdAt.IsZero() {
		co.createdAt = time.Now().UTC()
		event := events.NewComponentOriginsCreatedEvent(co.ID(), co.componentID, co.createdAt)
		co.apply(event)
		co.RaiseEvent(event)
	}

	if co.acquiredVia.IsEmpty() {
		event := events.NewAcquiredViaRelationshipSetEvent(co.ID(), co.componentID, entityID, notes, linkedAt)
		co.apply(event)
		co.RaiseEvent(event)
		return nil
	}

	if co.acquiredVia.EntityID() == entityID.String() {
		if co.acquiredVia.Notes().Equals(notes) {
			return nil
		}
		oldNotes := co.acquiredVia.Notes()
		event := events.NewAcquiredViaNotesUpdatedEvent(co.ID(), co.componentID, entityID, oldNotes, notes)
		co.apply(event)
		co.RaiseEvent(event)
		return nil
	}

	oldEntityID, _ := valueobjects.NewAcquiredEntityIDFromString(co.acquiredVia.EntityID())
	event := events.NewAcquiredViaRelationshipReplacedEvent(co.ID(), co.componentID, oldEntityID, entityID, notes, linkedAt)
	co.apply(event)
	co.RaiseEvent(event)
	return nil
}

func (co *ComponentOrigins) ClearAcquiredVia() error {
	if co.acquiredVia.IsEmpty() {
		return ErrNoAcquiredViaRelationship
	}

	entityID, _ := valueobjects.NewAcquiredEntityIDFromString(co.acquiredVia.EntityID())
	event := events.NewAcquiredViaRelationshipClearedEvent(co.ID(), co.componentID, entityID)
	co.apply(event)
	co.RaiseEvent(event)
	return nil
}

func (co *ComponentOrigins) SetPurchasedFrom(vendorID valueobjects.VendorID, notes valueobjects.Notes) error {
	linkedAt := time.Now().UTC()

	if co.createdAt.IsZero() {
		co.createdAt = time.Now().UTC()
		event := events.NewComponentOriginsCreatedEvent(co.ID(), co.componentID, co.createdAt)
		co.apply(event)
		co.RaiseEvent(event)
	}

	if co.purchasedFrom.IsEmpty() {
		event := events.NewPurchasedFromRelationshipSetEvent(co.ID(), co.componentID, vendorID, notes, linkedAt)
		co.apply(event)
		co.RaiseEvent(event)
		return nil
	}

	if co.purchasedFrom.EntityID() == vendorID.String() {
		if co.purchasedFrom.Notes().Equals(notes) {
			return nil
		}
		oldNotes := co.purchasedFrom.Notes()
		event := events.NewPurchasedFromNotesUpdatedEvent(co.ID(), co.componentID, vendorID, oldNotes, notes)
		co.apply(event)
		co.RaiseEvent(event)
		return nil
	}

	oldVendorID, _ := valueobjects.NewVendorIDFromString(co.purchasedFrom.EntityID())
	event := events.NewPurchasedFromRelationshipReplacedEvent(co.ID(), co.componentID, oldVendorID, vendorID, notes, linkedAt)
	co.apply(event)
	co.RaiseEvent(event)
	return nil
}

func (co *ComponentOrigins) ClearPurchasedFrom() error {
	if co.purchasedFrom.IsEmpty() {
		return ErrNoPurchasedFromRelationship
	}

	vendorID, _ := valueobjects.NewVendorIDFromString(co.purchasedFrom.EntityID())
	event := events.NewPurchasedFromRelationshipClearedEvent(co.ID(), co.componentID, vendorID)
	co.apply(event)
	co.RaiseEvent(event)
	return nil
}

func (co *ComponentOrigins) SetBuiltBy(teamID valueobjects.InternalTeamID, notes valueobjects.Notes) error {
	linkedAt := time.Now().UTC()

	if co.createdAt.IsZero() {
		co.createdAt = time.Now().UTC()
		event := events.NewComponentOriginsCreatedEvent(co.ID(), co.componentID, co.createdAt)
		co.apply(event)
		co.RaiseEvent(event)
	}

	if co.builtBy.IsEmpty() {
		event := events.NewBuiltByRelationshipSetEvent(co.ID(), co.componentID, teamID, notes, linkedAt)
		co.apply(event)
		co.RaiseEvent(event)
		return nil
	}

	if co.builtBy.EntityID() == teamID.String() {
		if co.builtBy.Notes().Equals(notes) {
			return nil
		}
		oldNotes := co.builtBy.Notes()
		event := events.NewBuiltByNotesUpdatedEvent(co.ID(), co.componentID, teamID, oldNotes, notes)
		co.apply(event)
		co.RaiseEvent(event)
		return nil
	}

	oldTeamID, _ := valueobjects.NewInternalTeamIDFromString(co.builtBy.EntityID())
	event := events.NewBuiltByRelationshipReplacedEvent(co.ID(), co.componentID, oldTeamID, teamID, notes, linkedAt)
	co.apply(event)
	co.RaiseEvent(event)
	return nil
}

func (co *ComponentOrigins) ClearBuiltBy() error {
	if co.builtBy.IsEmpty() {
		return ErrNoBuiltByRelationship
	}

	teamID, _ := valueobjects.NewInternalTeamIDFromString(co.builtBy.EntityID())
	event := events.NewBuiltByRelationshipClearedEvent(co.ID(), co.componentID, teamID)
	co.apply(event)
	co.RaiseEvent(event)
	return nil
}

func (co *ComponentOrigins) Delete() error {
	deletedAt := time.Now().UTC()
	event := events.NewComponentOriginsDeletedEvent(co.ID(), co.componentID, deletedAt)
	co.apply(event)
	co.RaiseEvent(event)
	return nil
}

func (co *ComponentOrigins) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ComponentOriginsCreated:
		co.AggregateRoot = domain.NewAggregateRootWithID(e.AggregateID())
		co.componentID, _ = valueobjects.NewComponentIDFromString(e.ComponentID)
		co.acquiredVia = valueobjects.EmptyOriginLink()
		co.purchasedFrom = valueobjects.EmptyOriginLink()
		co.builtBy = valueobjects.EmptyOriginLink()
		co.createdAt = e.CreatedAt
	case events.AcquiredViaRelationshipSet:
		notes, _ := valueobjects.NewNotes(e.Notes)
		co.acquiredVia = valueobjects.NewOriginLink(e.EntityID, notes, e.LinkedAt)
	case events.AcquiredViaRelationshipReplaced:
		notes, _ := valueobjects.NewNotes(e.Notes)
		co.acquiredVia = valueobjects.NewOriginLink(e.NewEntityID, notes, e.LinkedAt)
	case events.AcquiredViaNotesUpdated:
		notes, _ := valueobjects.NewNotes(e.NewNotes)
		co.acquiredVia = valueobjects.NewOriginLink(co.acquiredVia.EntityID(), notes, co.acquiredVia.LinkedAt())
	case events.AcquiredViaRelationshipCleared:
		co.acquiredVia = valueobjects.EmptyOriginLink()
	case events.PurchasedFromRelationshipSet:
		notes, _ := valueobjects.NewNotes(e.Notes)
		co.purchasedFrom = valueobjects.NewOriginLink(e.VendorID, notes, e.LinkedAt)
	case events.PurchasedFromRelationshipReplaced:
		notes, _ := valueobjects.NewNotes(e.Notes)
		co.purchasedFrom = valueobjects.NewOriginLink(e.NewVendorID, notes, e.LinkedAt)
	case events.PurchasedFromNotesUpdated:
		notes, _ := valueobjects.NewNotes(e.NewNotes)
		co.purchasedFrom = valueobjects.NewOriginLink(co.purchasedFrom.EntityID(), notes, co.purchasedFrom.LinkedAt())
	case events.PurchasedFromRelationshipCleared:
		co.purchasedFrom = valueobjects.EmptyOriginLink()
	case events.BuiltByRelationshipSet:
		notes, _ := valueobjects.NewNotes(e.Notes)
		co.builtBy = valueobjects.NewOriginLink(e.TeamID, notes, e.LinkedAt)
	case events.BuiltByRelationshipReplaced:
		notes, _ := valueobjects.NewNotes(e.Notes)
		co.builtBy = valueobjects.NewOriginLink(e.NewTeamID, notes, e.LinkedAt)
	case events.BuiltByNotesUpdated:
		notes, _ := valueobjects.NewNotes(e.NewNotes)
		co.builtBy = valueobjects.NewOriginLink(co.builtBy.EntityID(), notes, co.builtBy.LinkedAt())
	case events.BuiltByRelationshipCleared:
		co.builtBy = valueobjects.EmptyOriginLink()
	case events.ComponentOriginsDeleted:
		co.isDeleted = true
	}
}
