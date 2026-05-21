package commands

type PlacementInput struct {
	TargetBusinessDomainID string
	ResultingName          string
}

type CaptureDirection struct {
	EnterpriseCapabilityID string
	Type                   string
	SourceCapabilityIDs    []string
	Placements             []PlacementInput
	Horizon                string
	Narrative              string
}

func (c CaptureDirection) CommandName() string { return "CaptureDirection" }

type AdvanceDirection struct {
	DirectionID  string
	TargetStatus string
}

func (c AdvanceDirection) CommandName() string { return "AdvanceDirection" }

type RejectDirection struct {
	DirectionID string
}

func (c RejectDirection) CommandName() string { return "RejectDirection" }

type UpdateDirection struct {
	DirectionID         string
	Narrative           *string
	Horizon             *string
	SourceCapabilityIDs *[]string
	Placements          *[]PlacementInput
}

func (c UpdateDirection) CommandName() string { return "UpdateDirection" }
