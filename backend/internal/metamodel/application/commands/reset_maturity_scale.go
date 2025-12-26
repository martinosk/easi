package commands

type ResetMaturityScale struct {
	ID         string
	ModifiedBy string
}

func (c ResetMaturityScale) CommandName() string {
	return "ResetMaturityScale"
}
