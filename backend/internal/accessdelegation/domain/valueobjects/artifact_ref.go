package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidArtifactType = errors.New("invalid artifact type: must be capability, component, or view")
	ErrEmptyArtifactID     = errors.New("artifact ID cannot be empty")
)

type ArtifactType string

const (
	ArtifactTypeCapability ArtifactType = "capability"
	ArtifactTypeComponent  ArtifactType = "component"
	ArtifactTypeView       ArtifactType = "view"
)

var validArtifactTypes = map[string]ArtifactType{
	"capability": ArtifactTypeCapability,
	"component":  ArtifactTypeComponent,
	"view":       ArtifactTypeView,
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
