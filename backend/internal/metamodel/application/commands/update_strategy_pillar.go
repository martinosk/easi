package commands

type UpdateStrategyPillar struct {
	ConfigID        string
	PillarID        string
	Name            string
	Description     string
	ModifiedBy      string
	ExpectedVersion *int
}

func (c UpdateStrategyPillar) CommandName() string {
	return "UpdateStrategyPillar"
}
