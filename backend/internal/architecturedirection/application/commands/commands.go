package commands

type CaptureDirection struct {
	EnterpriseCapabilityID string
	Type                   string
	SourceCapabilityIDs    []string
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
	DirectionID string
	Narrative   *string
	Horizon     *string
}

func (c UpdateDirection) CommandName() string { return "UpdateDirection" }

type AddDirectionSource struct {
	DirectionID  string
	CapabilityID string
	Actor        string
}

func (c AddDirectionSource) CommandName() string { return "AddDirectionSource" }

type RemoveDirectionSource struct {
	DirectionID  string
	CapabilityID string
	Actor        string
}

func (c RemoveDirectionSource) CommandName() string { return "RemoveDirectionSource" }

type SetStandardApplication struct {
	EnterpriseCapabilityID string
	ApplicationID          string
	Narrative              string
}

func (c SetStandardApplication) CommandName() string { return "SetStandardApplication" }

type AssessRealization struct {
	CapabilityID string
	ComponentID  string
	Grade        string
	Rationale    string
	AssessedBy   string
}

func (c AssessRealization) CommandName() string { return "AssessRealization" }

type RemoveTimeAssessment struct {
	CapabilityID string
	ComponentID  string
	RemovedBy    string
}

func (c RemoveTimeAssessment) CommandName() string { return "RemoveTimeAssessment" }
