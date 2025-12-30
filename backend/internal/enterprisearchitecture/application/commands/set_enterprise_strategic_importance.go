package commands

type SetEnterpriseStrategicImportance struct {
	EnterpriseCapabilityID string
	PillarID               string
	PillarName             string
	Importance             int
	Rationale              string
	ID                     string
}

func (c SetEnterpriseStrategicImportance) CommandName() string {
	return "SetEnterpriseStrategicImportance"
}
