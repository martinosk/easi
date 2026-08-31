package commands

type UpdateEnterpriseCapability struct {
	ID          string
	Name        string
	Description string
	Category    string
}

func (c UpdateEnterpriseCapability) CommandName() string {
	return "UpdateEnterpriseCapability"
}
