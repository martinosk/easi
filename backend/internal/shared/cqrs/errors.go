package cqrs

import "errors"

var (
	// ErrInvalidCommand is returned when command type assertion fails
	ErrInvalidCommand = errors.New("invalid command type")

	// ErrInvalidQuery is returned when query type assertion fails
	ErrInvalidQuery = errors.New("invalid query type")
)
