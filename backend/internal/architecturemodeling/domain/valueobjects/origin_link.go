package valueobjects

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type OriginLink struct {
	entityID string
	notes    Notes
	linkedAt time.Time
}

func NewOriginLink(entityID string, notes Notes, linkedAt time.Time) OriginLink {
	return OriginLink{
		entityID: entityID,
		notes:    notes,
		linkedAt: linkedAt,
	}
}

func EmptyOriginLink() OriginLink {
	return OriginLink{
		entityID: "",
		notes:    Notes{},
		linkedAt: time.Time{},
	}
}

func (o OriginLink) EntityID() string {
	return o.entityID
}

func (o OriginLink) Notes() Notes {
	return o.notes
}

func (o OriginLink) LinkedAt() time.Time {
	return o.linkedAt
}

func (o OriginLink) IsEmpty() bool {
	return o.entityID == ""
}

func (o OriginLink) Equals(other domain.ValueObject) bool {
	if otherLink, ok := other.(OriginLink); ok {
		return o.entityID == otherLink.entityID &&
			o.notes.Equals(otherLink.notes) &&
			o.linkedAt.Equal(otherLink.linkedAt)
	}
	return false
}
