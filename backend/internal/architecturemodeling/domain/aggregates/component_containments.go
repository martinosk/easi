package aggregates

import (
	"errors"
	"fmt"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var (
	ErrSelfContainment                   = errors.New("a component cannot be part of itself")
	ErrPartAlreadyAttached               = errors.New("component is already part of another component")
	ErrPartHasParts                      = errors.New("a component that has parts cannot become a part")
	ErrParentIsPart                      = errors.New("a component that is a part cannot accept parts")
	ErrNotAPart                          = errors.New("component is not part of any component")
	ErrUnknownComponentContainmentsEvent = errors.New("unknown event type for component containments aggregate")
)

type Containment struct {
	Part   valueobjects.ComponentID
	Parent valueobjects.ComponentID
	Kind   valueobjects.ContainmentKind
}

type ComponentContainments struct {
	domain.AggregateRoot
	parents map[string]Containment
}

func NewComponentContainments() *ComponentContainments {
	return &ComponentContainments{
		AggregateRoot: domain.NewAggregateRootWithID(sharedvo.NewUUIDValue().Value()),
		parents:       map[string]Containment{},
	}
}

func LoadComponentContainmentsFromHistory(history []domain.DomainEvent) (*ComponentContainments, error) {
	aggregate := &ComponentContainments{
		AggregateRoot: domain.NewAggregateRoot(),
		parents:       map[string]Containment{},
	}
	var applyErr error
	aggregate.LoadFromHistory(history, func(event domain.DomainEvent) {
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

func (c *ComponentContainments) Attach(part, parent valueobjects.ComponentID, kind valueobjects.ContainmentKind) error {
	if err := c.checkAttachable(part, parent); err != nil {
		return err
	}
	c.raise(events.NewComponentAttached(c.ID(), part.Value(), parent.Value(), kind.String()))
	return nil
}

func (c *ComponentContainments) checkAttachable(part, parent valueobjects.ComponentID) error {
	switch {
	case part.Equals(parent):
		return ErrSelfContainment
	case c.IsPart(part):
		return ErrPartAlreadyAttached
	case c.HasParts(part):
		return ErrPartHasParts
	case c.IsPart(parent):
		return ErrParentIsPart
	}
	return nil
}

func (c *ComponentContainments) Detach(part valueobjects.ComponentID) error {
	containment, ok := c.parents[part.Value()]
	if !ok {
		return ErrNotAPart
	}
	c.raise(events.NewComponentDetached(c.ID(), part.Value(), containment.Parent.Value()))
	return nil
}

func (c *ComponentContainments) ContainmentOf(part valueobjects.ComponentID) (Containment, bool) {
	containment, ok := c.parents[part.Value()]
	return containment, ok
}

func (c *ComponentContainments) IsPart(component valueobjects.ComponentID) bool {
	_, ok := c.parents[component.Value()]
	return ok
}

func (c *ComponentContainments) HasParts(component valueobjects.ComponentID) bool {
	for _, containment := range c.parents {
		if containment.Parent.Equals(component) {
			return true
		}
	}
	return false
}

func (c *ComponentContainments) raise(event domain.DomainEvent) {
	if err := c.apply(event); err != nil {
		panic(fmt.Sprintf("architecturemodeling: in-process apply failed: %v", err))
	}
	c.RaiseEvent(event)
}

func (c *ComponentContainments) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.ComponentAttached:
		return c.applyAttached(e)
	case events.ComponentDetached:
		return c.applyDetached(e)
	default:
		return fmt.Errorf("%w: %T", ErrUnknownComponentContainmentsEvent, event)
	}
}

func (c *ComponentContainments) applyAttached(e events.ComponentAttached) error {
	containment, err := containmentFromEvent(e.PartID, e.ParentID, e.Kind)
	if err != nil {
		return err
	}
	c.adoptIdentity(e.ContainmentsID)
	c.parents[e.PartID] = containment
	return nil
}

func (c *ComponentContainments) applyDetached(e events.ComponentDetached) error {
	c.adoptIdentity(e.ContainmentsID)
	delete(c.parents, e.PartID)
	return nil
}

func (c *ComponentContainments) adoptIdentity(id string) {
	if c.ID() != id {
		c.AggregateRoot = domain.NewAggregateRootWithID(id)
	}
}

func containmentFromEvent(partID, parentID, kind string) (Containment, error) {
	part, err := valueobjects.NewComponentIDFromString(partID)
	if err != nil {
		return Containment{}, fmt.Errorf("%w: part id %q: %v", domain.ErrCorruptedEvent, partID, err)
	}
	parent, err := valueobjects.NewComponentIDFromString(parentID)
	if err != nil {
		return Containment{}, fmt.Errorf("%w: parent id %q: %v", domain.ErrCorruptedEvent, parentID, err)
	}
	containmentKind, err := valueobjects.NewContainmentKind(kind)
	if err != nil {
		return Containment{}, fmt.Errorf("%w: containment kind %q: %v", domain.ErrCorruptedEvent, kind, err)
	}
	return Containment{Part: part, Parent: parent, Kind: containmentKind}, nil
}
