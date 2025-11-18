package commands

type UpdateViewEdgeType struct {
	ViewID   string
	EdgeType string
}

func (c UpdateViewEdgeType) CommandName() string {
	return "UpdateViewEdgeType"
}
