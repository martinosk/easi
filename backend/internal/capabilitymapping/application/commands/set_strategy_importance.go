package commands

type SetStrategyImportance struct {
	BusinessDomainID string
	CapabilityID     string
	PillarID         string
	Importance       int
	Rationale        string
}

func (c SetStrategyImportance) CommandName() string {
	return "SetStrategyImportance"
}
