package contracts

import "time"

type CapabilityCreatedPayload struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentID    string    `json:"parentId"`
	Level       string    `json:"level"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CapabilityUpdatedPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CapabilityDeletedPayload struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type CapabilityMetadataUpdatedPayload struct {
	ID             string `json:"id"`
	StrategyPillar string `json:"strategyPillar"`
	PillarWeight   int    `json:"pillarWeight"`
	MaturityValue  int    `json:"maturityValue"`
	OwnershipModel string `json:"ownershipModel"`
	PrimaryOwner   string `json:"primaryOwner"`
	EAOwner        string `json:"eaOwner"`
	Status         string `json:"status"`
}

type CapabilityParentChangedPayload struct {
	CapabilityID string    `json:"capabilityId"`
	OldParentID  string    `json:"oldParentId"`
	NewParentID  string    `json:"newParentId"`
	OldLevel     string    `json:"oldLevel"`
	NewLevel     string    `json:"newLevel"`
	Timestamp    time.Time `json:"timestamp"`
}

type CapabilityLevelChangedPayload struct {
	CapabilityID string    `json:"capabilityId"`
	OldLevel     string    `json:"oldLevel"`
	NewLevel     string    `json:"newLevel"`
	Timestamp    time.Time `json:"timestamp"`
}

type CapabilityAssignedToDomainPayload struct {
	ID               string    `json:"id"`
	BusinessDomainID string    `json:"businessDomainId"`
	CapabilityID     string    `json:"capabilityId"`
	AssignedAt       time.Time `json:"assignedAt"`
}

type CapabilityUnassignedFromDomainPayload struct {
	ID               string    `json:"id"`
	BusinessDomainID string    `json:"businessDomainId"`
	CapabilityID     string    `json:"capabilityId"`
	UnassignedAt     time.Time `json:"unassignedAt"`
}

type SystemLinkedToCapabilityPayload struct {
	ID               string    `json:"id"`
	CapabilityID     string    `json:"capabilityId"`
	ComponentID      string    `json:"componentId"`
	ComponentName    string    `json:"componentName"`
	RealizationLevel string    `json:"realizationLevel"`
	Notes            string    `json:"notes"`
	LinkedAt         time.Time `json:"linkedAt"`
}

type SystemRealizationDeletedPayload struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type ApplicationFitScoreSetPayload struct {
	ID          string    `json:"id"`
	ComponentID string    `json:"componentId"`
	PillarID    string    `json:"pillarId"`
	PillarName  string    `json:"pillarName"`
	Score       int       `json:"score"`
	Rationale   string    `json:"rationale"`
	ScoredAt    time.Time `json:"scoredAt"`
	ScoredBy    string    `json:"scoredBy"`
}

type ApplicationFitScoreRemovedPayload struct {
	ID          string    `json:"id"`
	ComponentID string    `json:"componentId"`
	PillarID    string    `json:"pillarId"`
	RemovedAt   time.Time `json:"removedAt"`
	RemovedBy   string    `json:"removedBy"`
}

type EffectiveImportanceRecalculatedPayload struct {
	CapabilityID     string `json:"capabilityId"`
	BusinessDomainID string `json:"businessDomainId"`
	PillarID         string `json:"pillarId"`
	Importance       int    `json:"importance"`
}
