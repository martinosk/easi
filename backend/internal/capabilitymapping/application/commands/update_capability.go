package commands

type UpdateCapability struct {
	ID          string
	Name        string
	Description string
}

func (c UpdateCapability) CommandName() string {
	return "UpdateCapability"
}
