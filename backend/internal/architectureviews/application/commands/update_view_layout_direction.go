package commands

type UpdateViewLayoutDirection struct {
	ViewID          string
	LayoutDirection string
}

func (c UpdateViewLayoutDirection) CommandName() string {
	return "UpdateViewLayoutDirection"
}
