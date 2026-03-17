package commands

type CascadeDeleteCapability struct {
	ID                          string
	Cascade                     bool
	DeleteRealisingApplications bool
}

func (c CascadeDeleteCapability) CommandName() string {
	return "CascadeDeleteCapability"
}
