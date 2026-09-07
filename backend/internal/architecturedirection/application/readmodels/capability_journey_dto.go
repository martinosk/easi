package readmodels

import (
	"time"

	"easi/backend/internal/shared/types"
)

type TargetPeriodDTO struct {
	Year    int `json:"year"`
	Quarter int `json:"quarter"`
}

type JourneyApplicationRefDTO struct {
	ComponentID   string `json:"componentId"`
	ComponentName string `json:"componentName"`
	Stale         bool   `json:"stale"`
}

type JourneyMoveDTO struct {
	TargetDomainID    string `json:"targetDomainId"`
	TargetDomainName  string `json:"targetDomainName"`
	TargetDomainStale bool   `json:"targetDomainStale"`
	TargetParentID    string `json:"targetParentId"`
	TargetParentName  string `json:"targetParentName"`
	TargetParentStale bool   `json:"targetParentStale"`
	ResultingName     string `json:"resultingName"`
}

type JourneyMaturityDTO struct {
	TargetMaturity  int `json:"targetMaturity"`
	CurrentMaturity int `json:"currentMaturity"`
	MaturityGap     int `json:"maturityGap"`
}

type CapabilityJourneyMilestoneDTO struct {
	ID           string           `json:"id"`
	Label        string           `json:"label"`
	TargetPeriod *TargetPeriodDTO `json:"targetPeriod"`
	Status       string           `json:"status"`
	Links        types.Links      `json:"_links,omitempty"`
}

type CapabilityJourneyDTO struct {
	ID               string                          `json:"id"`
	CapabilityID     string                          `json:"capabilityId"`
	CapabilityName   string                          `json:"capabilityName"`
	CapabilityStale  bool                            `json:"capabilityStale"`
	Kind             string                          `json:"kind"`
	Status           string                          `json:"status"`
	Progress         *int                            `json:"progress"`
	TargetPeriod     *TargetPeriodDTO                `json:"targetPeriod"`
	Note             string                          `json:"note"`
	FromApplications []JourneyApplicationRefDTO      `json:"fromApplications"`
	ToApplication    JourneyApplicationRefDTO        `json:"toApplication"`
	Move             *JourneyMoveDTO                 `json:"move,omitempty"`
	Maturity         *JourneyMaturityDTO             `json:"maturity,omitempty"`
	Milestones       []CapabilityJourneyMilestoneDTO `json:"milestones"`
	PlannedBy        string                          `json:"plannedBy"`
	PlannedByName    string                          `json:"plannedByName"`
	PlannedAt        time.Time                       `json:"plannedAt"`
	UpdatedAt        *time.Time                      `json:"updatedAt,omitempty"`
	StartedAt        *time.Time                      `json:"startedAt"`
	CompletedAt      *time.Time                      `json:"completedAt"`
	AbandonedAt      *time.Time                      `json:"abandonedAt"`
	Links            types.Links                     `json:"_links,omitempty"`
}

type InsertJourneyParams struct {
	ID               string
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
	PlannedAt        time.Time
}

type JourneyTimestampColumn string

const (
	JourneyTimestampStarted   JourneyTimestampColumn = "started_at"
	JourneyTimestampCompleted JourneyTimestampColumn = "completed_at"
	JourneyTimestampAbandoned JourneyTimestampColumn = "abandoned_at"
)

var journeyTimestampColumns = map[JourneyTimestampColumn]struct{}{
	JourneyTimestampStarted:   {},
	JourneyTimestampCompleted: {},
	JourneyTimestampAbandoned: {},
}

type UpdateJourneyStatusParams struct {
	JourneyID  string
	Status     string
	Column     JourneyTimestampColumn
	OccurredAt time.Time
}

type UpdateJourneyDetailsParams struct {
	JourneyID     string
	Note          string
	TargetYear    *int
	TargetQuarter *int
	ResultingName string
}

type JourneyMilestoneUpsertParams struct {
	JourneyID     string
	MilestoneID   string
	Label         string
	TargetYear    *int
	TargetQuarter *int
	Status        string
	UpdatedAt     time.Time
}
