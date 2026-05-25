package commands

type CascadeDeleteCapability struct {
	ID      string
	Cascade bool
}

func (c CascadeDeleteCapability) CommandName() string {
	return "CascadeDeleteCapability"
}
