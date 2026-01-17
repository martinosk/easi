package commands

type UpdateVendor struct {
	ID                    string
	Name                  string
	ImplementationPartner string
	Notes                 string
}

func (c UpdateVendor) CommandName() string {
	return "UpdateVendor"
}
