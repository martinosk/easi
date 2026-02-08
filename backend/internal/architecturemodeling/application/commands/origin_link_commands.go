package commands

type SetOriginLink struct {
	ComponentID string
	OriginType  string
	EntityID    string
	Notes       string
}

func (c SetOriginLink) CommandName() string {
	return "SetOriginLink"
}

type ClearOriginLink struct {
	ComponentID string
	OriginType  string
}

func (c ClearOriginLink) CommandName() string {
	return "ClearOriginLink"
}
