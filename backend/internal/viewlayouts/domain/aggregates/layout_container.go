package aggregates

import (
	"time"

	"easi/backend/internal/viewlayouts/domain/valueobjects"
)

type LayoutContainer struct {
	id          valueobjects.LayoutContainerID
	contextType valueobjects.LayoutContextType
	contextRef  valueobjects.ContextRef
	preferences valueobjects.LayoutPreferences
	elements    map[string]valueobjects.ElementPosition
	version     int
	createdAt   time.Time
	updatedAt   time.Time
}

func NewLayoutContainer(
	contextType valueobjects.LayoutContextType,
	contextRef valueobjects.ContextRef,
	preferences valueobjects.LayoutPreferences,
) (*LayoutContainer, error) {
	now := time.Now().UTC()

	return &LayoutContainer{
		id:          valueobjects.NewLayoutContainerID(),
		contextType: contextType,
		contextRef:  contextRef,
		preferences: preferences,
		elements:    make(map[string]valueobjects.ElementPosition),
		version:     1,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func NewLayoutContainerWithState(
	id valueobjects.LayoutContainerID,
	contextType valueobjects.LayoutContextType,
	contextRef valueobjects.ContextRef,
	preferences valueobjects.LayoutPreferences,
	version int,
	createdAt, updatedAt time.Time,
) *LayoutContainer {
	return &LayoutContainer{
		id:          id,
		contextType: contextType,
		contextRef:  contextRef,
		preferences: preferences,
		elements:    make(map[string]valueobjects.ElementPosition),
		version:     version,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (c *LayoutContainer) ID() valueobjects.LayoutContainerID {
	return c.id
}

func (c *LayoutContainer) ContextType() valueobjects.LayoutContextType {
	return c.contextType
}

func (c *LayoutContainer) ContextRef() valueobjects.ContextRef {
	return c.contextRef
}

func (c *LayoutContainer) Preferences() valueobjects.LayoutPreferences {
	return c.preferences
}

func (c *LayoutContainer) Version() int {
	return c.version
}

func (c *LayoutContainer) CreatedAt() time.Time {
	return c.createdAt
}

func (c *LayoutContainer) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *LayoutContainer) Elements() []valueobjects.ElementPosition {
	elements := make([]valueobjects.ElementPosition, 0, len(c.elements))
	for _, e := range c.elements {
		elements = append(elements, e)
	}
	return elements
}

func (c *LayoutContainer) GetElement(elementID valueobjects.ElementID) *valueobjects.ElementPosition {
	if elem, exists := c.elements[elementID.Value()]; exists {
		return &elem
	}
	return nil
}

func (c *LayoutContainer) UpsertElement(position valueobjects.ElementPosition) error {
	c.elements[position.ElementID().Value()] = position
	return nil
}

func (c *LayoutContainer) RemoveElement(elementID valueobjects.ElementID) error {
	delete(c.elements, elementID.Value())
	return nil
}

func (c *LayoutContainer) UpdatePreferences(preferences valueobjects.LayoutPreferences) error {
	c.preferences = preferences
	return nil
}

func (c *LayoutContainer) IncrementVersion() {
	c.version++
}

func (c *LayoutContainer) SetElements(elements []valueobjects.ElementPosition) {
	c.elements = make(map[string]valueobjects.ElementPosition, len(elements))
	for _, e := range elements {
		c.elements[e.ElementID().Value()] = e
	}
}
