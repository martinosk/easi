package aggregates

import (
	"errors"
	"fmt"
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
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

	if err := aggregate.apply(event); err != nil {
		return nil, err
	}
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadComponentRelationFromHistory(events []domain.DomainEvent) (*ComponentRelation, error) {
	aggregate := &ComponentRelation{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	var applyErr error
	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		if applyErr != nil {
			return
		}
		applyErr = aggregate.apply(event)
	})
	if applyErr != nil {
		return nil, applyErr
	}

	return aggregate, nil
}

func (c *ComponentRelation) Update(name, description valueobjects.Description) error {
	event := events.NewComponentRelationUpdated(
		c.ID(),
		name.Value(),
		description.Value(),
	)

	if err := c.apply(event); err != nil {
		return err
	}
	c.RaiseEvent(event)

	return nil
}

func (c *ComponentRelation) Delete() error {
	event := events.NewComponentRelationDeleted(
		c.ID(),
		c.sourceComponentID.Value(),
		c.targetComponentID.Value(),
	)

	if err := c.apply(event); err != nil {
		return err
	}
	c.RaiseEvent(event)

	return nil
}

func (c *ComponentRelation) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.ComponentRelationCreated:
		return c.applyCreated(e)
	case events.ComponentRelationUpdated:
		return c.applyUpdated(e)
	case events.ComponentRelationDeleted:
		c.isDeleted = true
	}
	return nil
}

func (c *ComponentRelation) applyCreated(e events.ComponentRelationCreated) error {
	sourceID, err := valueobjects.NewComponentIDFromString(e.SourceComponentID)
	if err != nil {
		return fmt.Errorf("%w: source component ID %q: %v", domain.ErrCorruptedEvent, e.SourceComponentID, err)
	}
	targetID, err := valueobjects.NewComponentIDFromString(e.TargetComponentID)
	if err != nil {
		return fmt.Errorf("%w: target component ID %q: %v", domain.ErrCorruptedEvent, e.TargetComponentID, err)
	}
	relType, err := valueobjects.NewRelationType(e.RelationType)
	if err != nil {
		return fmt.Errorf("%w: relation type %q: %v", domain.ErrCorruptedEvent, e.RelationType, err)
	}
	name, err := valueobjects.NewDescription(e.Name)
	if err != nil {
		return fmt.Errorf("%w: name: %v", domain.ErrCorruptedEvent, err)
	}
	description, err := valueobjects.NewDescription(e.Description)
	if err != nil {
		return fmt.Errorf("%w: description: %v", domain.ErrCorruptedEvent, err)
	}
	c.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	c.sourceComponentID = sourceID
	c.targetComponentID = targetID
	c.relationType = relType
	c.name = name
	c.description = description
	c.createdAt = e.CreatedAt
	return nil
}

func (c *ComponentRelation) applyUpdated(e events.ComponentRelationUpdated) error {
	name, err := valueobjects.NewDescription(e.Name)
	if err != nil {
		return fmt.Errorf("%w: name: %v", domain.ErrCorruptedEvent, err)
	}
	description, err := valueobjects.NewDescription(e.Description)
	if err != nil {
		return fmt.Errorf("%w: description: %v", domain.ErrCorruptedEvent, err)
	}
	c.name = name
	c.description = description
	return nil
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
