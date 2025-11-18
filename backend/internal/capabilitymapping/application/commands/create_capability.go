package commands

type CreateCapability struct {
	Name        string
	Description string
	ParentID    string
	Level       string
	ID          string
}

func (c CreateCapability) CommandName() string {
	return "CreateCapability"
}
