package valueobjects

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type Expert struct {
	name    ExpertName
	role    ExpertRole
	contact ContactInfo
	addedAt time.Time
}

func NewExpert(name, role, contact string, addedAt time.Time) (Expert, error) {
	expertName, err := NewExpertName(name)
	if err != nil {
		return Expert{}, err
	}

	expertRole, err := NewExpertRole(role)
	if err != nil {
		return Expert{}, err
	}

	contactInfo, err := NewContactInfo(contact)
	if err != nil {
		return Expert{}, err
	}

	return Expert{
		name:    expertName,
		role:    expertRole,
		contact: contactInfo,
		addedAt: addedAt,
	}, nil
}

func MustNewExpert(name, role, contact string, addedAt time.Time) Expert {
	expert, err := NewExpert(name, role, contact, addedAt)
	if err != nil {
		panic(err)
	}
	return expert
}

func (e Expert) Name() ExpertName {
	return e.name
}

func (e Expert) Role() ExpertRole {
	return e.role
}

func (e Expert) Contact() ContactInfo {
	return e.contact
}

func (e Expert) AddedAt() time.Time {
	return e.addedAt
}

func (e Expert) Equals(other domain.ValueObject) bool {
	if otherExpert, ok := other.(Expert); ok {
		return e.name.Equals(otherExpert.name) &&
			e.role.Equals(otherExpert.role) &&
			e.contact.Equals(otherExpert.contact)
	}
	return false
}

func (e Expert) MatchesValues(name, role, contact string) bool {
	return e.name.Value() == name &&
		e.role.Value() == role &&
		e.contact.Value() == contact
}
