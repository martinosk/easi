package commands

type DeleteValueStream struct {
	ID string
}

func (c DeleteValueStream) CommandName() string {
	return "DeleteValueStream"
}
