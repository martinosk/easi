package commands

type AttachComponent struct {
	ComponentID string
	ParentID    string
	Kind        string
}

func (c AttachComponent) CommandName() string {
	return "AttachComponent"
}

type DetachComponent struct {
	ComponentID string
}

func (c DetachComponent) CommandName() string {
	return "DetachComponent"
}
