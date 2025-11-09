package domain

import "errors"

var (
	// ErrEmptyValue is returned when a required value is empty
	ErrEmptyValue = errors.New("value cannot be empty")

	// ErrInvalidValue is returned when a value doesn't meet validation criteria
	ErrInvalidValue = errors.New("invalid value")
)

// ValueObject is a marker interface for all value objects
type ValueObject interface {
	Equals(other ValueObject) bool
}
