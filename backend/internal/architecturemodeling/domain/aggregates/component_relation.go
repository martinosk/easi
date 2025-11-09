package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/domain"
)

var (
	// ErrSelfReference is returned when source and target are the same
	ErrSelfReference = errors.New("component cannot have a relation to itself")
)

// ComponentRelation represents a relation between two application components
type ComponentRelation struct {
	domain.AggregateRoot
	sourceComponentID valueobjects.ComponentID
	targetComponentID valueobjects.ComponentID
	relationType      valueobjects.RelationType
	name              valueobjects.Description
	description       valueobjects.Description
	createdAt         time.Time
}

// NewComponentRelation creates a new component relation
func NewComponentRelation(
	sourceID, targetID valueobjects.ComponentID,
	relationType valueobjects.RelationType,
	name, description valueobjects.Description,
) (*ComponentRelation, error) {
	// Validate no self-reference
	if sourceID.Equals(targetID) {
		return nil, ErrSelfReference
	}

	aggregate := &ComponentRelation{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	// Raise creation event
	event := events.NewComponentRelationCreated(
		aggregate.ID(),
		sourceID.Value(),
		targetID.Value(),
		relationType.Value(),
		name.Value(),
		description.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

// LoadFromHistory reconstructs the aggregate from events
func LoadComponentRelationFromHistory(events []domain.DomainEvent) (*ComponentRelation, error) {
	aggregate := &ComponentRelation{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

// apply applies an event to the aggregate
// Note: This method should NOT increment the version - that's handled by LoadFromHistory or RaiseEvent
func (c *ComponentRelation) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ComponentRelationCreated:
		c.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		c.sourceComponentID, _ = valueobjects.NewComponentIDFromString(e.SourceComponentID)
		c.targetComponentID, _ = valueobjects.NewComponentIDFromString(e.TargetComponentID)
		c.relationType, _ = valueobjects.NewRelationType(e.RelationType)
		c.name = valueobjects.NewDescription(e.Name)
		c.description = valueobjects.NewDescription(e.Description)
		c.createdAt = e.CreatedAt
	}
}

// SourceComponentID returns the source component ID
func (c *ComponentRelation) SourceComponentID() valueobjects.ComponentID {
	return c.sourceComponentID
}

// TargetComponentID returns the target component ID
func (c *ComponentRelation) TargetComponentID() valueobjects.ComponentID {
	return c.targetComponentID
}

// RelationType returns the relation type
func (c *ComponentRelation) RelationType() valueobjects.RelationType {
	return c.relationType
}

// Name returns the relation name
func (c *ComponentRelation) Name() valueobjects.Description {
	return c.name
}

// Description returns the relation description
func (c *ComponentRelation) Description() valueobjects.Description {
	return c.description
}

// CreatedAt returns when the relation was created
func (c *ComponentRelation) CreatedAt() time.Time {
	return c.createdAt
}
