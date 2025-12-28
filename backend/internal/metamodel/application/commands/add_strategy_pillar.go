package commands

type AddStrategyPillar struct {
	ConfigID    string
	Name        string
	Description string
	ModifiedBy  string
}

func (c AddStrategyPillar) CommandName() string {
	return "AddStrategyPillar"
}
