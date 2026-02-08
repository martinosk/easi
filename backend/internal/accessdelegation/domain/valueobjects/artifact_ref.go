package valueobjects

import (
	"errors"
	"regexp"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidArtifactType = errors.New("invalid artifact type: must be capability, component, view, domain, vendor, internal_team, or acquired_entity")
	ErrEmptyArtifactID     = errors.New("artifact ID cannot be empty")
	ErrInvalidArtifactID   = errors.New("artifact ID must be a valid UUID")
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type ArtifactType string

const (
	ArtifactTypeCapability      ArtifactType = "capability"
	ArtifactTypeComponent       ArtifactType = "component"
	ArtifactTypeView            ArtifactType = "view"
	ArtifactTypeDomain          ArtifactType = "domain"
	ArtifactTypeVendor          ArtifactType = "vendor"
	ArtifactTypeInternalTeam    ArtifactType = "internal_team"
	ArtifactTypeAcquiredEntity  ArtifactType = "acquired_entity"
)

var validArtifactTypes = map[string]ArtifactType{
	"capability":       ArtifactTypeCapability,
	"component":        ArtifactTypeComponent,
	"view":             ArtifactTypeView,
	"domain":           ArtifactTypeDomain,
	"vendor":           ArtifactTypeVendor,
	"internal_team":    ArtifactTypeInternalTeam,
	"acquired_entity":  ArtifactTypeAcquiredEntity,
}

func NewArtifactType(s string) (ArtifactType, error) {
	if at, ok := validArtifactTypes[s]; ok {
		return at, nil
	}
	return "", ErrInvalidArtifactType
}

func (at ArtifactType) String() string {
	return string(at)
}

type ArtifactRef struct {
	artifactType ArtifactType
	artifactID   string
}

func NewArtifactRef(artifactType ArtifactType, artifactID string) (ArtifactRef, error) {
	if artifactID == "" {
		return ArtifactRef{}, ErrEmptyArtifactID
	}
	if !uuidRegex.MatchString(artifactID) {
		return ArtifactRef{}, ErrInvalidArtifactID
	}
	return ArtifactRef{artifactType: artifactType, artifactID: artifactID}, nil
}

func (r ArtifactRef) Type() ArtifactType {
	return r.artifactType
}

func (r ArtifactRef) ID() string {
	return r.artifactID
}

func (r ArtifactRef) Equals(other domain.ValueObject) bool {
	otherRef, ok := other.(ArtifactRef)
	if !ok {
		return false
	}
	return r.artifactType == otherRef.artifactType && r.artifactID == otherRef.artifactID
}
