package commands

type CreateVendor struct {
	Name                  string
	ImplementationPartner string
	Notes                 string
}

func (c CreateVendor) CommandName() string {
	return "CreateVendor"
}
