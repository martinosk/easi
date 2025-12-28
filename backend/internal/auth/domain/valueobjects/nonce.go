package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"

	"golang.org/x/oauth2"
)

var ErrEmptyNonce = errors.New("nonce cannot be empty")

type Nonce struct {
	value string
}

func NewNonce() Nonce {
	return Nonce{value: oauth2.GenerateVerifier()}
}

func NonceFromValue(value string) (Nonce, error) {
	if value == "" {
		return Nonce{}, ErrEmptyNonce
	}
	return Nonce{value: value}, nil
}

func (n Nonce) Value() string {
	return n.value
}

func (n Nonce) Equals(other domain.ValueObject) bool {
	otherNonce, ok := other.(Nonce)
	if !ok {
		return false
	}
	return n.value == otherNonce.value
}
