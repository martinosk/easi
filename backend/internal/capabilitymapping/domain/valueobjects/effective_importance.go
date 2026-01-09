package valueobjects

import domain "easi/backend/internal/shared/eventsourcing"

type EffectiveImportance struct {
	importance         Importance
	sourceCapabilityID CapabilityID
	isInherited        bool
}

func NewEffectiveImportance(importance Importance, sourceCapabilityID CapabilityID, isInherited bool) EffectiveImportance {
	return EffectiveImportance{
		importance:         importance,
		sourceCapabilityID: sourceCapabilityID,
		isInherited:        isInherited,
	}
}

func (e EffectiveImportance) Importance() Importance {
	return e.importance
}

func (e EffectiveImportance) SourceCapabilityID() CapabilityID {
	return e.sourceCapabilityID
}

func (e EffectiveImportance) IsInherited() bool {
	return e.isInherited
}

func (e EffectiveImportance) Equals(other domain.ValueObject) bool {
	if otherEI, ok := other.(EffectiveImportance); ok {
		return e.importance.Equals(otherEI.importance) &&
			e.sourceCapabilityID.Equals(otherEI.sourceCapabilityID) &&
			e.isInherited == otherEI.isInherited
	}
	return false
}
