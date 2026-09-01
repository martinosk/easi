package commands

type ClassifyApplicationHosting struct {
	ComponentID string
	Hosting     string
}

func (c ClassifyApplicationHosting) CommandName() string {
	return "ClassifyApplicationHosting"
}
