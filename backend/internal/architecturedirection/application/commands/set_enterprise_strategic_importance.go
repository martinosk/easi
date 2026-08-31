package commands

type SetEnterpriseStrategicImportance struct {
	EnterpriseCapabilityID string
	PillarID               string
	PillarName             string
	Importance             int
	Rationale              string
}

func (c SetEnterpriseStrategicImportance) CommandName() string {
	return "SetEnterpriseStrategicImportance"
}
