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
	return co.setOriginRelationship(
		co.acquiredVia,
		entityID.String(),
		notes,
		func(linkedAt time.Time) domain.DomainEvent {
			return events.NewAcquiredViaRelationshipSetEvent(co.ID(), co.componentID, entityID, notes, linkedAt)
		},
		func(oldNotes valueobjects.Notes) domain.DomainEvent {
			return events.NewAcquiredViaNotesUpdatedEvent(co.ID(), co.componentID, entityID, oldNotes, notes)
		},
		func(linkedAt time.Time) domain.DomainEvent {
			oldEntityID, _ := valueobjects.NewAcquiredEntityIDFromString(co.acquiredVia.EntityID())
			return events.NewAcquiredViaRelationshipReplacedEvent(co.ID(), co.componentID, oldEntityID, entityID, notes, linkedAt)
		},
	)
}

func (co *ComponentOrigins) ClearAcquiredVia() error {
	return co.clearOriginRelationship(
		co.acquiredVia,
		ErrNoAcquiredViaRelationship,
		func() domain.DomainEvent {
			entityID, _ := valueobjects.NewAcquiredEntityIDFromString(co.acquiredVia.EntityID())
			return events.NewAcquiredViaRelationshipClearedEvent(co.ID(), co.componentID, entityID)
		},
	)
}

func (co *ComponentOrigins) SetPurchasedFrom(vendorID valueobjects.VendorID, notes valueobjects.Notes) error {
	return co.setOriginRelationship(
		co.purchasedFrom,
		vendorID.String(),
		notes,
		func(linkedAt time.Time) domain.DomainEvent {
			return events.NewPurchasedFromRelationshipSetEvent(co.ID(), co.componentID, vendorID, notes, linkedAt)
		},
		func(oldNotes valueobjects.Notes) domain.DomainEvent {
			return events.NewPurchasedFromNotesUpdatedEvent(co.ID(), co.componentID, vendorID, oldNotes, notes)
		},
		func(linkedAt time.Time) domain.DomainEvent {
			oldVendorID, _ := valueobjects.NewVendorIDFromString(co.purchasedFrom.EntityID())
			return events.NewPurchasedFromRelationshipReplacedEvent(co.ID(), co.componentID, oldVendorID, vendorID, notes, linkedAt)
		},
	)
}

func (co *ComponentOrigins) ClearPurchasedFrom() error {
	return co.clearOriginRelationship(
		co.purchasedFrom,
		ErrNoPurchasedFromRelationship,
		func() domain.DomainEvent {
			vendorID, _ := valueobjects.NewVendorIDFromString(co.purchasedFrom.EntityID())
			return events.NewPurchasedFromRelationshipClearedEvent(co.ID(), co.componentID, vendorID)
		},
	)
}

func (co *ComponentOrigins) SetBuiltBy(teamID valueobjects.InternalTeamID, notes valueobjects.Notes) error {
	return co.setOriginRelationship(
		co.builtBy,
		teamID.String(),
		notes,
		func(linkedAt time.Time) domain.DomainEvent {
			return events.NewBuiltByRelationshipSetEvent(co.ID(), co.componentID, teamID, notes, linkedAt)
		},
		func(oldNotes valueobjects.Notes) domain.DomainEvent {
			return events.NewBuiltByNotesUpdatedEvent(co.ID(), co.componentID, teamID, oldNotes, notes)
		},
		func(linkedAt time.Time) domain.DomainEvent {
			oldTeamID, _ := valueobjects.NewInternalTeamIDFromString(co.builtBy.EntityID())
			return events.NewBuiltByRelationshipReplacedEvent(co.ID(), co.componentID, oldTeamID, teamID, notes, linkedAt)
		},
	)
}

func (co *ComponentOrigins) ClearBuiltBy() error {
	return co.clearOriginRelationship(
		co.builtBy,
		ErrNoBuiltByRelationship,
		func() domain.DomainEvent {
			teamID, _ := valueobjects.NewInternalTeamIDFromString(co.builtBy.EntityID())
			return events.NewBuiltByRelationshipClearedEvent(co.ID(), co.componentID, teamID)
		},
	)
}

func (co *ComponentOrigins) Delete() error {
	deletedAt := time.Now().UTC()
	event := events.NewComponentOriginsDeletedEvent(co.ID(), co.componentID, deletedAt)
	co.applyAndRaise(event)
	return nil
}

func (co *ComponentOrigins) setOriginRelationship(
	current valueobjects.OriginLink,
	entityID string,
	notes valueobjects.Notes,
	newSetEvent func(time.Time) domain.DomainEvent,
	newNotesUpdatedEvent func(valueobjects.Notes) domain.DomainEvent,
	newReplacedEvent func(time.Time) domain.DomainEvent,
) error {
	co.ensureCreated()
	linkedAt := time.Now().UTC()

	if current.IsEmpty() {
		co.applyAndRaise(newSetEvent(linkedAt))
		return nil
	}

	if current.EntityID() == entityID {
		if current.Notes().Equals(notes) {
			return nil
		}
		co.applyAndRaise(newNotesUpdatedEvent(current.Notes()))
		return nil
	}

	co.applyAndRaise(newReplacedEvent(linkedAt))
	return nil
}

func (co *ComponentOrigins) clearOriginRelationship(
	current valueobjects.OriginLink,
	emptyErr error,
	newClearedEvent func() domain.DomainEvent,
) error {
	if current.IsEmpty() {
		return emptyErr
	}

	co.applyAndRaise(newClearedEvent())
	return nil
}

func (co *ComponentOrigins) ensureCreated() {
	if co.createdAt.IsZero() {
		co.createdAt = time.Now().UTC()
		event := events.NewComponentOriginsCreatedEvent(co.ID(), co.componentID, co.createdAt)
		co.applyAndRaise(event)
	}
}

func (co *ComponentOrigins) applyAndRaise(event domain.DomainEvent) {
	co.apply(event)
	co.RaiseEvent(event)
}

func (co *ComponentOrigins) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ComponentOriginsCreated:
		co.applyCreated(e)
	case events.ComponentOriginsDeleted:
		co.isDeleted = true
	case events.AcquiredViaRelationshipSet:
		co.acquiredVia = newOriginLink(e.EntityID, e.Notes, e.LinkedAt)
	case events.AcquiredViaRelationshipReplaced:
		co.acquiredVia = newOriginLink(e.NewEntityID, e.Notes, e.LinkedAt)
	case events.AcquiredViaNotesUpdated:
		co.acquiredVia = updatedOriginLink(co.acquiredVia, e.NewNotes)
	case events.AcquiredViaRelationshipCleared:
		co.acquiredVia = valueobjects.EmptyOriginLink()
	case events.PurchasedFromRelationshipSet:
		co.purchasedFrom = newOriginLink(e.VendorID, e.Notes, e.LinkedAt)
	case events.PurchasedFromRelationshipReplaced:
		co.purchasedFrom = newOriginLink(e.NewVendorID, e.Notes, e.LinkedAt)
	case events.PurchasedFromNotesUpdated:
		co.purchasedFrom = updatedOriginLink(co.purchasedFrom, e.NewNotes)
	case events.PurchasedFromRelationshipCleared:
		co.purchasedFrom = valueobjects.EmptyOriginLink()
	case events.BuiltByRelationshipSet:
		co.builtBy = newOriginLink(e.TeamID, e.Notes, e.LinkedAt)
	case events.BuiltByRelationshipReplaced:
		co.builtBy = newOriginLink(e.NewTeamID, e.Notes, e.LinkedAt)
	case events.BuiltByNotesUpdated:
		co.builtBy = updatedOriginLink(co.builtBy, e.NewNotes)
	case events.BuiltByRelationshipCleared:
		co.builtBy = valueobjects.EmptyOriginLink()
	}
}

func (co *ComponentOrigins) applyCreated(e events.ComponentOriginsCreated) {
	co.AggregateRoot = domain.NewAggregateRootWithID(e.AggregateID())
	co.componentID, _ = valueobjects.NewComponentIDFromString(e.ComponentID)
	co.acquiredVia = valueobjects.EmptyOriginLink()
	co.purchasedFrom = valueobjects.EmptyOriginLink()
	co.builtBy = valueobjects.EmptyOriginLink()
	co.createdAt = e.CreatedAt
}

func newOriginLink(entityID string, notesStr string, linkedAt time.Time) valueobjects.OriginLink {
	notes, _ := valueobjects.NewNotes(notesStr)
	return valueobjects.NewOriginLink(entityID, notes, linkedAt)
}

func updatedOriginLink(current valueobjects.OriginLink, notesStr string) valueobjects.OriginLink {
	notes, _ := valueobjects.NewNotes(notesStr)
	return valueobjects.NewOriginLink(current.EntityID(), notes, current.LinkedAt())
}
