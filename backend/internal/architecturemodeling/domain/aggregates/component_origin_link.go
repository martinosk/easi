package aggregates

import (
	"errors"
	"log"
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrNoOriginLink      = errors.New("no origin link exists")
	ErrOriginLinkDeleted = errors.New("origin link has been deleted")
)

type ComponentOriginLink struct {
	domain.AggregateRoot
	componentID valueobjects.ComponentID
	originType  valueobjects.OriginType
	link        valueobjects.OriginLink
	createdAt   time.Time
	isDeleted   bool
}

func BuildOriginLinkAggregateID(originType, componentID string) string {
	return "origin-link:" + originType + ":" + componentID
}

func NewComponentOriginLink(componentID valueobjects.ComponentID, originType valueobjects.OriginType) (*ComponentOriginLink, error) {
	aggregateID := BuildOriginLinkAggregateID(originType.String(), componentID.String())
	ol := &ComponentOriginLink{
		AggregateRoot: domain.NewAggregateRootWithID(aggregateID),
	}
	base := events.NewOriginLinkBase(aggregateID, componentID.String(), originType.String())
	event := events.NewOriginLinkCreatedEvent(base, time.Now().UTC())
	ol.apply(event)
	ol.RaiseEvent(event)
	return ol, nil
}

func LoadComponentOriginLinkFromHistory(domainEvents []domain.DomainEvent) (*ComponentOriginLink, error) {
	ol := &ComponentOriginLink{
		AggregateRoot: domain.NewAggregateRoot(),
	}
	ol.LoadFromHistory(domainEvents, func(event domain.DomainEvent) {
		ol.apply(event)
	})
	return ol, nil
}

func (ol *ComponentOriginLink) ComponentID() valueobjects.ComponentID {
	return ol.componentID
}

func (ol *ComponentOriginLink) OriginType() valueobjects.OriginType {
	return ol.originType
}

func (ol *ComponentOriginLink) Link() valueobjects.OriginLink {
	return ol.link
}

func (ol *ComponentOriginLink) CreatedAt() time.Time {
	return ol.createdAt
}

func (ol *ComponentOriginLink) IsDeleted() bool {
	return ol.isDeleted
}

func (ol *ComponentOriginLink) Set(entityID string, notes valueobjects.Notes) error {
	if ol.isDeleted {
		return ErrOriginLinkDeleted
	}
	ol.ensureCreated()
	linkedAt := time.Now().UTC()
	base := ol.eventBase()

	if ol.link.IsEmpty() {
		ol.applyAndRaise(events.NewOriginLinkSetEvent(base, entityID, notes.String(), linkedAt))
		return nil
	}

	if ol.link.EntityID() == entityID {
		if ol.link.Notes().Equals(notes) {
			return nil
		}
		ol.applyAndRaise(events.NewOriginLinkNotesUpdatedEvent(base, entityID, ol.link.Notes().String(), notes.String()))
		return nil
	}

	ol.applyAndRaise(events.NewOriginLinkReplacedEvent(base, events.OriginLinkReplacement{
		OldEntityID: ol.link.EntityID(), NewEntityID: entityID, Notes: notes.String(), LinkedAt: linkedAt,
	}))
	return nil
}

func (ol *ComponentOriginLink) Clear() error {
	if ol.isDeleted {
		return ErrOriginLinkDeleted
	}
	if ol.link.IsEmpty() {
		return ErrNoOriginLink
	}
	ol.applyAndRaise(events.NewOriginLinkClearedEvent(ol.eventBase(), ol.link.EntityID()))
	return nil
}

func (ol *ComponentOriginLink) Delete() error {
	deletedAt := time.Now().UTC()
	event := events.NewOriginLinkDeletedEvent(ol.eventBase(), deletedAt)
	ol.applyAndRaise(event)
	return nil
}

func (ol *ComponentOriginLink) ensureCreated() {
	if ol.createdAt.IsZero() {
		ol.createdAt = time.Now().UTC()
		event := events.NewOriginLinkCreatedEvent(ol.eventBase(), ol.createdAt)
		ol.applyAndRaise(event)
	}
}

func (ol *ComponentOriginLink) eventBase() events.OriginLinkBase {
	return events.NewOriginLinkBase(ol.ID(), ol.componentID.String(), ol.originType.String())
}

func (ol *ComponentOriginLink) applyAndRaise(event domain.DomainEvent) {
	ol.apply(event)
	ol.RaiseEvent(event)
}

func (ol *ComponentOriginLink) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.OriginLinkCreated:
		ol.applyCreated(e)
	case events.OriginLinkSet:
		ol.link = newOriginLink(e.EntityID, e.Notes, e.LinkedAt)
	case events.OriginLinkReplaced:
		ol.link = newOriginLink(e.NewEntityID, e.Notes, e.LinkedAt)
	case events.OriginLinkNotesUpdated:
		ol.link = updatedOriginLink(ol.link, e.NewNotes)
	case events.OriginLinkCleared:
		ol.link = valueobjects.EmptyOriginLink()
	case events.OriginLinkDeleted:
		ol.isDeleted = true
	}
}

func (ol *ComponentOriginLink) applyCreated(e events.OriginLinkCreated) {
	ol.AggregateRoot = domain.NewAggregateRootWithID(e.AggregateID())
	componentID, err := valueobjects.NewComponentIDFromString(e.ComponentID)
	if err != nil {
		log.Printf("corrupted event: invalid componentID %q in OriginLinkCreated: %v", e.ComponentID, err)
	}
	originType, err := valueobjects.NewOriginType(e.OriginType)
	if err != nil {
		log.Printf("corrupted event: invalid originType %q in OriginLinkCreated: %v", e.OriginType, err)
	}
	ol.componentID = componentID
	ol.originType = originType
	ol.link = valueobjects.EmptyOriginLink()
	ol.createdAt = e.CreatedAt
}

func newOriginLink(entityID string, notesStr string, linkedAt time.Time) valueobjects.OriginLink {
	notes, err := valueobjects.NewNotes(notesStr)
	if err != nil {
		log.Printf("corrupted event: invalid notes in origin link event: %v", err)
	}
	return valueobjects.NewOriginLink(entityID, notes, linkedAt)
}

func updatedOriginLink(current valueobjects.OriginLink, notesStr string) valueobjects.OriginLink {
	notes, err := valueobjects.NewNotes(notesStr)
	if err != nil {
		log.Printf("corrupted event: invalid notes in origin link event: %v", err)
	}
	return valueobjects.NewOriginLink(current.EntityID(), notes, current.LinkedAt())
}
