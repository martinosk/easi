package commands

type CreateValueStream struct {
	Name        string
	Description string
}

func (c CreateValueStream) CommandName() string {
	return "CreateValueStream"
}
