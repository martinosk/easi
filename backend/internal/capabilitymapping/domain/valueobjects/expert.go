package valueobjects

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
	name    string
	role    string
	contact string
	addedAt time.Time
}

func NewExpert(name, role, contact string, addedAt time.Time) (Expert, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Expert{}, ErrExpertNameEmpty
	}

	role = strings.TrimSpace(role)
	if role == "" {
		return Expert{}, ErrExpertRoleEmpty
	}

	contact = strings.TrimSpace(contact)
	if contact == "" {
		return Expert{}, ErrExpertContactEmpty
	}

	return Expert{
		name:    name,
		role:    role,
		contact: contact,
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

func (e Expert) Name() string {
	return e.name
}

func (e Expert) Role() string {
	return e.role
}

func (e Expert) Contact() string {
	return e.contact
}

func (e Expert) AddedAt() time.Time {
	return e.addedAt
}

func (e Expert) MatchesValues(name, role, contact string) bool {
	return e.name == name &&
		e.role == role &&
		e.contact == contact
}
