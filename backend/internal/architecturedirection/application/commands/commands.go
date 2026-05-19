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
	DirectionID string
	TargetStatus string // "proposed" | "agreed"
}

func (c AdvanceDirection) CommandName() string { return "AdvanceDirection" }

type RejectDirection struct {
	DirectionID string
}

func (c RejectDirection) CommandName() string { return "RejectDirection" }

type UpdateDirectionNarrative struct {
	DirectionID string
	Narrative   string
}

func (c UpdateDirectionNarrative) CommandName() string { return "UpdateDirectionNarrative" }

type UpdateDirectionHorizon struct {
	DirectionID string
	Horizon     string
}

func (c UpdateDirectionHorizon) CommandName() string { return "UpdateDirectionHorizon" }

type UpdateDirectionSourceCapabilities struct {
	DirectionID         string
	SourceCapabilityIDs []string
}

func (c UpdateDirectionSourceCapabilities) CommandName() string { return "UpdateDirectionSourceCapabilities" }

type UpdateDirectionPlacements struct {
	DirectionID string
	Placements  []PlacementInput
}

func (c UpdateDirectionPlacements) CommandName() string { return "UpdateDirectionPlacements" }
