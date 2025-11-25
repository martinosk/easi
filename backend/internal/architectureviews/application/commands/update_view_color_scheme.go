package commands

type UpdateViewColorScheme struct {
	ViewID      string
	ColorScheme string
}

func (c UpdateViewColorScheme) CommandName() string {
	return "UpdateViewColorScheme"
}
