package valueobjects

type FitConfigurationParams struct {
	enabled  bool
	criteria FitCriteria
	fitType  FitType
}

func NewFitConfigurationParams(enabled bool, criteria FitCriteria, fitType FitType) FitConfigurationParams {
	return FitConfigurationParams{
		enabled:  enabled,
		criteria: criteria,
		fitType:  fitType,
	}
}

func (f FitConfigurationParams) Enabled() bool {
	return f.enabled
}

func (f FitConfigurationParams) Criteria() FitCriteria {
	return f.criteria
}

func (f FitConfigurationParams) FitType() FitType {
	return f.fitType
}
