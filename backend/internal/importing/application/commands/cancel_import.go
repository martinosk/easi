package commands

type CancelImport struct {
	ID string
}

func (c CancelImport) CommandName() string {
	return "CancelImport"
}
