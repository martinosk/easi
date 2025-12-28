package commands

type RemoveStrategyPillar struct {
	ConfigID   string
	PillarID   string
	ModifiedBy string
}

func (c RemoveStrategyPillar) CommandName() string {
	return "RemoveStrategyPillar"
}
