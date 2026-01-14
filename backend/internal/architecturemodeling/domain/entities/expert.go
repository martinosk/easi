package entities

import (
	"time"

	"easi/backend/internal/architecturemodeling/domain/valueobjects"
)

type Expert struct {
	name    valueobjects.ExpertName
	role    valueobjects.ExpertRole
	contact valueobjects.ContactInfo
	addedAt time.Time
}

func NewExpert(name, role, contact string) (*Expert, error) {
	expertName, err := valueobjects.NewExpertName(name)
	if err != nil {
		return nil, err
	}

	expertRole, err := valueobjects.NewExpertRole(role)
	if err != nil {
		return nil, err
	}

	contactInfo, err := valueobjects.NewContactInfo(contact)
	if err != nil {
		return nil, err
	}

	return &Expert{
		name:    expertName,
		role:    expertRole,
		contact: contactInfo,
		addedAt: time.Now().UTC(),
	}, nil
}

func NewExpertWithAddedAt(name, role, contact string, addedAt time.Time) (*Expert, error) {
	expertName, err := valueobjects.NewExpertName(name)
	if err != nil {
		return nil, err
	}

	expertRole, err := valueobjects.NewExpertRole(role)
	if err != nil {
		return nil, err
	}

	contactInfo, err := valueobjects.NewContactInfo(contact)
	if err != nil {
		return nil, err
	}

	return &Expert{
		name:    expertName,
		role:    expertRole,
		contact: contactInfo,
		addedAt: addedAt,
	}, nil
}

func (e *Expert) Name() valueobjects.ExpertName {
	return e.name
}

func (e *Expert) Role() valueobjects.ExpertRole {
	return e.role
}

func (e *Expert) Contact() valueobjects.ContactInfo {
	return e.contact
}

func (e *Expert) AddedAt() time.Time {
	return e.addedAt
}
