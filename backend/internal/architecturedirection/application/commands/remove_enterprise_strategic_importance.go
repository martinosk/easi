package commands

type RemoveEnterpriseStrategicImportance struct {
	ID string
}

func (c RemoveEnterpriseStrategicImportance) CommandName() string {
	return "RemoveEnterpriseStrategicImportance"
}
