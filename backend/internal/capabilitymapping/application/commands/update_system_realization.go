package commands

type UpdateSystemRealization struct {
	ID               string
	RealizationLevel string
	Notes            string
}

func (c UpdateSystemRealization) CommandName() string {
	return "UpdateSystemRealization"
}
