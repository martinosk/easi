package valueobjects

import "errors"

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
