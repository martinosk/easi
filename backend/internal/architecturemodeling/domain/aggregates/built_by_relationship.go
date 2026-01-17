package aggregates

import (
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type BuiltByRelationship struct {
	domain.AggregateRoot
	internalTeamID valueobjects.InternalTeamID
	componentID    valueobjects.ComponentID
	notes          valueobjects.Notes
	createdAt      time.Time
	isDeleted      bool
}

func NewBuiltByRelationship(
	internalTeamID valueobjects.InternalTeamID,
	componentID valueobjects.ComponentID,
	notes valueobjects.Notes,
) (*BuiltByRelationship, error) {
	aggregate := &BuiltByRelationship{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewBuiltByRelationshipCreated(
		aggregate.ID(),
		internalTeamID.Value(),
		componentID.Value(),
		notes.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadBuiltByRelationshipFromHistory(events []domain.DomainEvent) (*BuiltByRelationship, error) {
	aggregate := &BuiltByRelationship{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (b *BuiltByRelationship) Delete() error {
	event := events.NewBuiltByRelationshipDeleted(
		b.ID(),
		b.internalTeamID.Value(),
		b.componentID.Value(),
	)

	b.apply(event)
	b.RaiseEvent(event)

	return nil
}

func (b *BuiltByRelationship) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.BuiltByRelationshipCreated:
		b.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		b.internalTeamID, _ = valueobjects.NewInternalTeamIDFromString(e.InternalTeamID)
		b.componentID, _ = valueobjects.NewComponentIDFromString(e.ComponentID)
		b.notes = valueobjects.MustNewNotes(e.Notes)
		b.createdAt = e.CreatedAt
	case events.BuiltByRelationshipDeleted:
		b.isDeleted = true
	}
}

func (b *BuiltByRelationship) InternalTeamID() valueobjects.InternalTeamID {
	return b.internalTeamID
}

func (b *BuiltByRelationship) ComponentID() valueobjects.ComponentID {
	return b.componentID
}

func (b *BuiltByRelationship) Notes() valueobjects.Notes {
	return b.notes
}

func (b *BuiltByRelationship) CreatedAt() time.Time {
	return b.createdAt
}

func (b *BuiltByRelationship) IsDeleted() bool {
	return b.isDeleted
}
