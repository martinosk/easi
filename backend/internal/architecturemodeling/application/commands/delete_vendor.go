package commands

type DeleteVendor struct {
	ID string
}

func (c DeleteVendor) CommandName() string {
	return "DeleteVendor"
}
