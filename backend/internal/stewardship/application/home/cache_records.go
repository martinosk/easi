package home

import "time"

type CapabilityID string

type DomainID string

type ComponentID string

type RealizationID string

type EditGrantID string

type CachedAssignment struct {
	CapabilityID CapabilityID
	DomainID     DomainID
}

type CachedPlacement struct {
	CapabilityID CapabilityID
	ParentID     CapabilityID
	Level        string
}

type CachedCapabilityMetadata struct {
	CapabilityID CapabilityID
	Status       string
	EAOwner      string
}

type CachedCapability struct {
	ID       CapabilityID
	Name     string
	Level    string
	ParentID CapabilityID
}

type CachedRealization struct {
	ID           RealizationID
	CapabilityID CapabilityID
	ComponentID  ComponentID
}

const (
	ownershipOwned     = "owned"
	ownershipNominated = "nominated"
	ownershipUnknown   = "unknown"
)

type CachedOwnership struct {
	ComponentID ComponentID
	State       string
	OwnerKind   string
	OwnerID     string
}

type CachedTimeAssessment struct {
	CapabilityID CapabilityID
	ComponentID  ComponentID
	Grade        string
	AssessedAt   time.Time
}

type CachedEditGrant struct {
	ID           EditGrantID
	ArtifactType string
	ArtifactID   string
	GranteeEmail string
	ExpiresAt    time.Time
}
