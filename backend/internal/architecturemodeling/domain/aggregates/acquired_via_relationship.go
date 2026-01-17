package aggregates

import (
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type AcquiredViaRelationship struct {
	domain.AggregateRoot
	acquiredEntityID valueobjects.AcquiredEntityID
	componentID      valueobjects.ComponentID
	notes            valueobjects.Notes
	createdAt        time.Time
	isDeleted        bool
}

func NewAcquiredViaRelationship(
	acquiredEntityID valueobjects.AcquiredEntityID,
	componentID valueobjects.ComponentID,
	notes valueobjects.Notes,
) (*AcquiredViaRelationship, error) {
	aggregate := &AcquiredViaRelationship{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewAcquiredViaRelationshipCreated(
		aggregate.ID(),
		acquiredEntityID.Value(),
		componentID.Value(),
		notes.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadAcquiredViaRelationshipFromHistory(events []domain.DomainEvent) (*AcquiredViaRelationship, error) {
	aggregate := &AcquiredViaRelationship{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (a *AcquiredViaRelationship) Delete() error {
	event := events.NewAcquiredViaRelationshipDeleted(
		a.ID(),
		a.acquiredEntityID.Value(),
		a.componentID.Value(),
	)

	a.apply(event)
	a.RaiseEvent(event)

	return nil
}

func (a *AcquiredViaRelationship) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.AcquiredViaRelationshipCreated:
		a.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		a.acquiredEntityID, _ = valueobjects.NewAcquiredEntityIDFromString(e.AcquiredEntityID)
		a.componentID, _ = valueobjects.NewComponentIDFromString(e.ComponentID)
		a.notes = valueobjects.MustNewNotes(e.Notes)
		a.createdAt = e.CreatedAt
	case events.AcquiredViaRelationshipDeleted:
		a.isDeleted = true
	}
}

func (a *AcquiredViaRelationship) AcquiredEntityID() valueobjects.AcquiredEntityID {
	return a.acquiredEntityID
}

func (a *AcquiredViaRelationship) ComponentID() valueobjects.ComponentID {
	return a.componentID
}

func (a *AcquiredViaRelationship) Notes() valueobjects.Notes {
	return a.notes
}

func (a *AcquiredViaRelationship) CreatedAt() time.Time {
	return a.createdAt
}

func (a *AcquiredViaRelationship) IsDeleted() bool {
	return a.isDeleted
}
