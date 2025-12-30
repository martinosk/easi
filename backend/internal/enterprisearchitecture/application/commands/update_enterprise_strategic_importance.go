package commands

type UpdateEnterpriseStrategicImportance struct {
	ID         string
	Importance int
	Rationale  string
}

func (c UpdateEnterpriseStrategicImportance) CommandName() string {
	return "UpdateEnterpriseStrategicImportance"
}
