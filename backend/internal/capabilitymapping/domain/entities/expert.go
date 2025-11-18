package entities

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrExpertNameEmpty    = errors.New("expert name cannot be empty")
	ErrExpertRoleEmpty    = errors.New("expert role cannot be empty")
	ErrExpertContactEmpty = errors.New("expert contact cannot be empty")
)

type Expert struct {
	name      string
	role      string
	contact   string
	addedAt   time.Time
}

func NewExpert(name, role, contact string) (*Expert, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrExpertNameEmpty
	}

	role = strings.TrimSpace(role)
	if role == "" {
		return nil, ErrExpertRoleEmpty
	}

	contact = strings.TrimSpace(contact)
	if contact == "" {
		return nil, ErrExpertContactEmpty
	}

	return &Expert{
		name:    name,
		role:    role,
		contact: contact,
		addedAt: time.Now().UTC(),
	}, nil
}

func (e *Expert) Name() string {
	return e.name
}

func (e *Expert) Role() string {
	return e.role
}

func (e *Expert) Contact() string {
	return e.contact
}

func (e *Expert) AddedAt() time.Time {
	return e.addedAt
}
