package valueobjects

// RelationProperties encapsulates the properties of a component relation
type RelationProperties struct {
	sourceID     ComponentID
	targetID     ComponentID
	relationType RelationType
	name         Description
	description  Description
}

type RelationPropertiesParams struct {
	SourceID     ComponentID
	TargetID     ComponentID
	RelationType RelationType
	Name         Description
	Description  Description
}

func NewRelationProperties(params RelationPropertiesParams) RelationProperties {
	return RelationProperties{
		sourceID:     params.SourceID,
		targetID:     params.TargetID,
		relationType: params.RelationType,
		name:         params.Name,
		description:  params.Description,
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
