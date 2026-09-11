package publishedlanguage

const (
	MetaModelConfigurationCreated = "MetaModelConfigurationCreated"
	StrategyPillarAdded           = "StrategyPillarAdded"
	StrategyPillarUpdated         = "StrategyPillarUpdated"
	StrategyPillarRemoved         = "StrategyPillarRemoved"
	PillarFitConfigurationUpdated = "PillarFitConfigurationUpdated"
	MaturityScaleConfigUpdated    = "MaturityScaleConfigUpdated"
	MaturityScaleConfigReset      = "MaturityScaleConfigReset"

	SubjectAttributeDefined       = "SubjectAttributeDefined"
	SubjectAttributeRenamed       = "SubjectAttributeRenamed"
	SubjectAttributeRetired       = "SubjectAttributeRetired"
	SubjectAttributeReactivated   = "SubjectAttributeReactivated"
	SubjectAttributeOptionAdded   = "SubjectAttributeOptionAdded"
	SubjectAttributeOptionRetired = "SubjectAttributeOptionRetired"
	SubjectAttributeBoundsChanged = "SubjectAttributeBoundsChanged"
)

func SubjectAttributeEventTypes() []string {
	return []string{
		SubjectAttributeDefined,
		SubjectAttributeRenamed,
		SubjectAttributeRetired,
		SubjectAttributeReactivated,
		SubjectAttributeOptionAdded,
		SubjectAttributeOptionRetired,
		SubjectAttributeBoundsChanged,
	}
}
