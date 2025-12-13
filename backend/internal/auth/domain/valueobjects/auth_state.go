package valueobjects

import (
	"errors"

	"golang.org/x/oauth2"

	"easi/backend/internal/shared/domain"
)

var ErrEmptyAuthState = errors.New("auth state cannot be empty")

type AuthState struct {
	value string
}

func NewAuthState() AuthState {
	return AuthState{value: oauth2.GenerateVerifier()}
}

func AuthStateFromValue(value string) (AuthState, error) {
	if value == "" {
		return AuthState{}, ErrEmptyAuthState
	}
	return AuthState{value: value}, nil
}

func (s AuthState) Value() string {
	return s.value
}

func (s AuthState) Equals(other domain.ValueObject) bool {
	otherState, ok := other.(AuthState)
	if !ok {
		return false
	}
	return s.value == otherState.value
}
