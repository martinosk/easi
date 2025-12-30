package commands

type CreateEnterpriseCapability struct {
	Name        string
	Description string
	Category    string
	ID          string
}

func (c CreateEnterpriseCapability) CommandName() string {
	return "CreateEnterpriseCapability"
}
