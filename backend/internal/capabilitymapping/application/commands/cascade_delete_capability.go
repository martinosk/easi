package commands

type CascadeDeleteCapability struct {
	ID                      string
	Cascade                 bool
	ConfirmedRealizationIDs []string
}

func (c CascadeDeleteCapability) CommandName() string {
	return "CascadeDeleteCapability"
}
