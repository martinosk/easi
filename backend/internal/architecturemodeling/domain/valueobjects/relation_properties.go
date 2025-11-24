package valueobjects

// RelationProperties encapsulates the properties of a component relation
type RelationProperties struct {
	sourceID     ComponentID
	targetID     ComponentID
	relationType RelationType
	name         Description
	description  Description
}

// NewRelationProperties creates a new relation properties value object
func NewRelationProperties(
	sourceID ComponentID,
	targetID ComponentID,
	relationType RelationType,
	name Description,
	description Description,
) RelationProperties {
	return RelationProperties{
		sourceID:     sourceID,
		targetID:     targetID,
		relationType: relationType,
		name:         name,
		description:  description,
	}
}

// SourceID returns the source component ID
func (p RelationProperties) SourceID() ComponentID {
	return p.sourceID
}

// TargetID returns the target component ID
func (p RelationProperties) TargetID() ComponentID {
	return p.targetID
}

// RelationType returns the relation type
func (p RelationProperties) RelationType() RelationType {
	return p.relationType
}

// Name returns the name
func (p RelationProperties) Name() Description {
	return p.name
}

// Description returns the description
func (p RelationProperties) Description() Description {
	return p.description
}