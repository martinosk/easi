package valueobjects

type ComponentMembership struct {
	components map[string]bool
}

func NewComponentMembership() ComponentMembership {
	return ComponentMembership{
		components: make(map[string]bool),
	}
}

func (c ComponentMembership) Add(componentID string) {
	c.components[componentID] = true
}

func (c ComponentMembership) Remove(componentID string) {
	delete(c.components, componentID)
}

func (c ComponentMembership) Contains(componentID string) bool {
	return c.components[componentID]
}

func (c ComponentMembership) ContainsString(componentID string) bool {
	return c.components[componentID]
}

func (c ComponentMembership) GetAll() []string {
	componentIDs := make([]string, 0, len(c.components))
	for id := range c.components {
		componentIDs = append(componentIDs, id)
	}
	return componentIDs
}

func (c ComponentMembership) Count() int {
	return len(c.components)
}

func (c ComponentMembership) IsEmpty() bool {
	return len(c.components) == 0
}