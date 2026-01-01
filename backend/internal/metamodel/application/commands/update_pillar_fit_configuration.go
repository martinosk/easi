package commands

type UpdatePillarFitConfiguration struct {
	ConfigID          string
	PillarID          string
	FitScoringEnabled bool
	FitCriteria       string
	ModifiedBy        string
	ExpectedVersion   *int
}

func (c UpdatePillarFitConfiguration) CommandName() string {
	return "UpdatePillarFitConfiguration"
}
