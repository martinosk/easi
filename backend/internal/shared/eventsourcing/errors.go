package domain

import "errors"

// Common domain errors
var (
	// ErrNotFound indicates a resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrDuplicateResource indicates a resource with the same identifier already exists
	ErrDuplicateResource = errors.New("duplicate resource")

	// ErrInvalidOperation indicates the operation is not allowed in the current state
	ErrInvalidOperation = errors.New("invalid operation")

	// ErrConflict indicates a business rule violation or conflict
	ErrConflict = errors.New("conflict")

	// ErrConcurrencyConflict indicates the resource was modified by another user
	ErrConcurrencyConflict = errors.New("resource was modified by another user")
)

// ValidationError represents a validation error with details
type ValidationError struct {
	Message string
	Field   string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}
