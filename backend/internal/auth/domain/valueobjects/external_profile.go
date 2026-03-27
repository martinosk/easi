package valueobjects

import domain "easi/backend/internal/shared/eventsourcing"

type ExternalProfile struct {
	name       string
	externalID string
}

func NewExternalProfile(name string, externalID string) ExternalProfile {
	return ExternalProfile{name: name, externalID: externalID}
}

func (p ExternalProfile) Name() string {
	return p.name
}

func (p ExternalProfile) ExternalID() string {
	return p.externalID
}

func (p ExternalProfile) HasName() bool {
	return p.name != ""
}

func (p ExternalProfile) HasExternalID() bool {
	return p.externalID != ""
}

func (p ExternalProfile) Equals(other domain.ValueObject) bool {
	otherProfile, ok := other.(ExternalProfile)
	if !ok {
		return false
	}
	return p.name == otherProfile.name && p.externalID == otherProfile.externalID
}
