package commands

type UpdateComponentRelation struct {
	ID          string
	Name        string
	Description string
}

func (c UpdateComponentRelation) CommandName() string {
	return "UpdateComponentRelation"
}
