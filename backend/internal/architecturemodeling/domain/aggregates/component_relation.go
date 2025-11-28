package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/domain"
)

var (
	ErrSelfReference = errors.New("component cannot have a relation to itself")
)

type ComponentRelation struct {
	domain.AggregateRoot
	sourceComponentID valueobjects.ComponentID
	targetComponentID valueobjects.ComponentID
	relationType      valueobjects.RelationType
	name              valueobjects.Description
	description       valueobjects.Description
	createdAt         time.Time
	isDeleted         bool
}

func NewComponentRelation(properties valueobjects.RelationProperties) (*ComponentRelation, error) {
	if properties.SourceID().Equals(properties.TargetID()) {
		return nil, ErrSelfReference
	}

	aggregate := &ComponentRelation{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewComponentRelationCreated(events.ComponentRelationParams{
		ID:          aggregate.ID(),
		SourceID:    properties.SourceID().Value(),
		TargetID:    properties.TargetID().Value(),
		Type:        properties.RelationType().Value(),
		Name:        properties.Name().Value(),
		Description: properties.Description().Value(),
	})

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadComponentRelationFromHistory(events []domain.DomainEvent) (*ComponentRelation, error) {
	aggregate := &ComponentRelation{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (c *ComponentRelation) Update(name, description valueobjects.Description) error {
	event := events.NewComponentRelationUpdated(
		c.ID(),
		name.Value(),
		description.Value(),
	)

	c.apply(event)
	c.RaiseEvent(event)

	return nil
}

func (c *ComponentRelation) Delete() error {
	event := events.NewComponentRelationDeleted(
		c.ID(),
		c.sourceComponentID.Value(),
		c.targetComponentID.Value(),
	)

	c.apply(event)
	c.RaiseEvent(event)

	return nil
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
	case events.ComponentRelationUpdated:
		c.name = valueobjects.NewDescription(e.Name)
		c.description = valueobjects.NewDescription(e.Description)
	case events.ComponentRelationDeleted:
		c.isDeleted = true
	}
}

func (c *ComponentRelation) SourceComponentID() valueobjects.ComponentID {
	return c.sourceComponentID
}

func (c *ComponentRelation) TargetComponentID() valueobjects.ComponentID {
	return c.targetComponentID
}

func (c *ComponentRelation) RelationType() valueobjects.RelationType {
	return c.relationType
}

func (c *ComponentRelation) Name() valueobjects.Description {
	return c.name
}

func (c *ComponentRelation) Description() valueobjects.Description {
	return c.description
}

func (c *ComponentRelation) CreatedAt() time.Time {
	return c.createdAt
}

func (c *ComponentRelation) IsDeleted() bool {
	return c.isDeleted
}
