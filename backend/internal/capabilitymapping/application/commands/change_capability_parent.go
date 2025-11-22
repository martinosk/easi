package commands

type ChangeCapabilityParent struct {
	CapabilityID string
	NewParentID  string
}

func (c ChangeCapabilityParent) CommandName() string {
	return "ChangeCapabilityParent"
}
