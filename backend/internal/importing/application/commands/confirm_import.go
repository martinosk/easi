package commands

type ConfirmImport struct {
	ID string
}

func (c ConfirmImport) CommandName() string {
	return "ConfirmImport"
}
