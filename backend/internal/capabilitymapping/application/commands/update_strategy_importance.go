package commands

type UpdateStrategyImportance struct {
	ImportanceID string
	Importance   int
	Rationale    string
}

func (c UpdateStrategyImportance) CommandName() string {
	return "UpdateStrategyImportance"
}
