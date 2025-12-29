package commands

type RemoveStrategyImportance struct {
	ImportanceID string
}

func (c RemoveStrategyImportance) CommandName() string {
	return "RemoveStrategyImportance"
}
