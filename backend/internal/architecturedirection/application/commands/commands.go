package commands

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

type AssignRealizationRole struct {
	CapabilityID string
	ComponentID  string
	Role         string
	AssignedBy   string
}

func (c AssignRealizationRole) CommandName() string { return "AssignRealizationRole" }

type ClearRealizationRole struct {
	CapabilityID string
	ComponentID  string
	ClearedBy    string
}

func (c ClearRealizationRole) CommandName() string { return "ClearRealizationRole" }

type PlanJourney struct {
	CapabilityID     string
	Kind             string
	FromComponentIDs []string
	ToComponentID    string
	Note             string
	TargetYear       *int
	TargetQuarter    *int
	TargetDomainID   string
	TargetParentID   string
	ResultingName    string
	TargetMaturity   *int
	PlannedBy        string
}

func (c PlanJourney) CommandName() string { return "PlanJourney" }

type StartJourney struct {
	JourneyID string
	Actor     string
}

func (c StartJourney) CommandName() string { return "StartJourney" }

type CompleteJourney struct {
	JourneyID string
	Actor     string
}

func (c CompleteJourney) CommandName() string { return "CompleteJourney" }

type AbandonJourney struct {
	JourneyID string
	Actor     string
}

func (c AbandonJourney) CommandName() string { return "AbandonJourney" }

type UpdateJourneyProgress struct {
	JourneyID string
	Progress  int
	Actor     string
}

func (c UpdateJourneyProgress) CommandName() string { return "UpdateJourneyProgress" }

type UpdateJourneyDetails struct {
	JourneyID     string
	Note          string
	TargetYear    *int
	TargetQuarter *int
	ResultingName string
	Actor         string
}

func (c UpdateJourneyDetails) CommandName() string { return "UpdateJourneyDetails" }

type ChangeJourneySourceApplications struct {
	JourneyID        string
	FromComponentIDs []string
	Actor            string
}

func (c ChangeJourneySourceApplications) CommandName() string {
	return "ChangeJourneySourceApplications"
}

type AddJourneyMilestone struct {
	JourneyID     string
	Label         string
	TargetYear    *int
	TargetQuarter *int
	Status        string
	Actor         string
}

func (c AddJourneyMilestone) CommandName() string { return "AddJourneyMilestone" }

type UpdateJourneyMilestone struct {
	JourneyID     string
	MilestoneID   string
	Label         string
	TargetYear    *int
	TargetQuarter *int
	Status        string
	Actor         string
}

func (c UpdateJourneyMilestone) CommandName() string { return "UpdateJourneyMilestone" }

type RemoveJourneyMilestone struct {
	JourneyID   string
	MilestoneID string
	Actor       string
}

func (c RemoveJourneyMilestone) CommandName() string { return "RemoveJourneyMilestone" }

type ReorderJourneyMilestones struct {
	JourneyID    string
	MilestoneIDs []string
	Actor        string
}

func (c ReorderJourneyMilestones) CommandName() string { return "ReorderJourneyMilestones" }
