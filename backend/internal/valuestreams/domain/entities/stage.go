package entities

import (
	"easi/backend/internal/valuestreams/domain/valueobjects"
)

type Stage struct {
	id             valueobjects.StageID
	name           valueobjects.StageName
	description    valueobjects.Description
	position       valueobjects.StagePosition
	capabilityRefs []valueobjects.CapabilityRef
}

func NewStage(
	id valueobjects.StageID,
	name valueobjects.StageName,
	description valueobjects.Description,
	position valueobjects.StagePosition,
) Stage {
	return Stage{
		id:             id,
		name:           name,
		description:    description,
		position:       position,
		capabilityRefs: []valueobjects.CapabilityRef{},
	}
}

func (s Stage) ID() valueobjects.StageID {
	return s.id
}

func (s Stage) Name() valueobjects.StageName {
	return s.name
}

func (s Stage) Description() valueobjects.Description {
	return s.description
}

func (s Stage) Position() valueobjects.StagePosition {
	return s.position
}

func (s Stage) CapabilityRefs() []valueobjects.CapabilityRef {
	refs := make([]valueobjects.CapabilityRef, len(s.capabilityRefs))
	copy(refs, s.capabilityRefs)
	return refs
}

func (s Stage) HasCapability(ref valueobjects.CapabilityRef) bool {
	for _, r := range s.capabilityRefs {
		if r.Equals(ref) {
			return true
		}
	}
	return false
}

func (s Stage) WithName(name valueobjects.StageName) Stage {
	s.name = name
	return s
}

func (s Stage) WithDescription(description valueobjects.Description) Stage {
	s.description = description
	return s
}

func (s Stage) WithPosition(position valueobjects.StagePosition) Stage {
	s.position = position
	return s
}

func (s Stage) WithAddedCapability(ref valueobjects.CapabilityRef) Stage {
	s.capabilityRefs = append(s.CapabilityRefs(), ref)
	return s
}

func (s Stage) WithRemovedCapability(ref valueobjects.CapabilityRef) Stage {
	refs := make([]valueobjects.CapabilityRef, 0, len(s.capabilityRefs))
	for _, r := range s.capabilityRefs {
		if !r.Equals(ref) {
			refs = append(refs, r)
		}
	}
	s.capabilityRefs = refs
	return s
}
