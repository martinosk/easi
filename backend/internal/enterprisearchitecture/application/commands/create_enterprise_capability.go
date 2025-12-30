package commands

type CreateEnterpriseCapability struct {
	Name        string
	Description string
	Category    string
}

func (c CreateEnterpriseCapability) CommandName() string {
	return "CreateEnterpriseCapability"
}
