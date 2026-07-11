package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/google/uuid"
)

var (
	ErrInvalidFieldID  = errors.New("field ID must be a valid UUID")
	ErrInvalidOptionID = errors.New("option ID must be a valid UUID")
)

type FieldID struct {
	value string
}

func NewFieldID() FieldID {
	return FieldID{value: uuid.New().String()}
}

func NewFieldIDFromString(value string) (FieldID, error) {
	if _, err := uuid.Parse(value); err != nil {
		return FieldID{}, ErrInvalidFieldID
	}
	return FieldID{value: value}, nil
}

func (f FieldID) Value() string {
	return f.value
}

func (f FieldID) Equals(other domain.ValueObject) bool {
	if o, ok := other.(FieldID); ok {
		return f.value == o.value
	}
	return false
}

func (f FieldID) String() string {
	return f.value
}

type OptionID struct {
	value string
}

func NewOptionID() OptionID {
	return OptionID{value: uuid.New().String()}
}

func NewOptionIDFromString(value string) (OptionID, error) {
	if _, err := uuid.Parse(value); err != nil {
		return OptionID{}, ErrInvalidOptionID
	}
	return OptionID{value: value}, nil
}

func (o OptionID) Value() string {
	return o.value
}

func (o OptionID) Equals(other domain.ValueObject) bool {
	if other2, ok := other.(OptionID); ok {
		return o.value == other2.value
	}
	return false
}

func (o OptionID) String() string {
	return o.value
}
