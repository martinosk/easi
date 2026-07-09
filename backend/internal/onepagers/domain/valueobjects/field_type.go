package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidFieldType = errors.New("invalid field type")

const fieldTypeSelection = "selection"

type FieldType struct {
	value string
}

var fieldTypeValues = []string{
	"text",
	"number",
	"date",
	"link",
	fieldTypeSelection,
	"contact-person",
}

func NewFieldType(value string) (FieldType, error) {
	for _, v := range fieldTypeValues {
		if v == value {
			return FieldType{value: value}, nil
		}
	}
	return FieldType{}, ErrInvalidFieldType
}

func (f FieldType) Value() string {
	return f.value
}

func (f FieldType) IsSelection() bool {
	return f.value == fieldTypeSelection
}

func (f FieldType) Equals(other domain.ValueObject) bool {
	if o, ok := other.(FieldType); ok {
		return f.value == o.value
	}
	return false
}

func (f FieldType) String() string {
	return f.value
}
