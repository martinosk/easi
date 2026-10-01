package home

import "time"

type CapabilityID string

type DomainID string

type CachedAssignment struct {
	CapabilityID CapabilityID
	DomainID     DomainID
}

type CachedPlacement struct {
	CapabilityID CapabilityID
	ParentID     string
	Level        string
}

type CachedCapabilityMetadata struct {
	CapabilityID CapabilityID
	Status       string
	EAOwner      string
}

type CachedCapability struct {
	ID       string
	Name     string
	Level    string
	ParentID string
}

type CachedRealization struct {
	ID           string
	CapabilityID string
	ComponentID  string
}

const (
	ownershipOwned     = "owned"
	ownershipNominated = "nominated"
	ownershipUnknown   = "unknown"
)

type CachedOwnership struct {
	ComponentID string
	State       string
	OwnerKind   string
	OwnerID     string
}

type CachedTimeAssessment struct {
	CapabilityID string
	ComponentID  string
	Grade        string
	AssessedAt   time.Time
}

type CachedEditGrant struct {
	ID           string
	ArtifactType string
	ArtifactID   string
	GranteeEmail string
	ExpiresAt    time.Time
}
