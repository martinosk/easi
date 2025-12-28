package commands

type PillarOperation string

const (
	PillarOperationAdd    PillarOperation = "add"
	PillarOperationUpdate PillarOperation = "update"
	PillarOperationRemove PillarOperation = "remove"
)

type PillarChange struct {
	Operation   PillarOperation
	PillarID    string
	Name        string
	Description string
}

type BatchUpdateStrategyPillars struct {
	ConfigID        string
	Changes         []PillarChange
	ModifiedBy      string
	ExpectedVersion *int
}

func (c BatchUpdateStrategyPillars) CommandName() string {
	return "BatchUpdateStrategyPillars"
}
