package commands

type UpdateValueStream struct {
	ID          string
	Name        string
	Description string
}

func (c UpdateValueStream) CommandName() string {
	return "UpdateValueStream"
}
