package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrGrantorIDEmpty    = errors.New("grantor ID must not be empty")
	ErrGrantorEmailEmpty = errors.New("grantor email must not be empty")
)

type Grantor struct {
	id    string
	email string
}

func NewGrantor(id, email string) (Grantor, error) {
	if id == "" {
		return Grantor{}, ErrGrantorIDEmpty
	}
	if email == "" {
		return Grantor{}, ErrGrantorEmailEmpty
	}
	return Grantor{id: id, email: email}, nil
}

func (g Grantor) ID() string    { return g.id }
func (g Grantor) Email() string { return g.email }

func (g Grantor) Equals(other domain.ValueObject) bool {
	otherGrantor, ok := other.(Grantor)
	if !ok {
		return false
	}
	return g.id == otherGrantor.id && g.email == otherGrantor.email
}
