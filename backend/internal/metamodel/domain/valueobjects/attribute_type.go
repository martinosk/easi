package valueobjects

import (
	"errors"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidAttributeType = errors.New("invalid attribute type")

const (
	attributeTypeSelection = "selection"
	attributeTypeNumber    = "number"
)

type AttributeType struct {
	value string
}

var attributeTypeValues = []string{"text", attributeTypeNumber, "date", "link", attributeTypeSelection, "contact-person"}

func NewAttributeType(value string) (AttributeType, error) {
	for _, v := range attributeTypeValues {
		if v == value {
			return AttributeType{value: value}, nil
		}
	}
	return AttributeType{}, ErrInvalidAttributeType
}

func (a AttributeType) Value() string {
	return a.value
}

func (a AttributeType) IsSelection() bool {
	return a.value == attributeTypeSelection
}

func (a AttributeType) IsNumber() bool {
	return a.value == attributeTypeNumber
}

func (a AttributeType) Equals(other domain.ValueObject) bool {
	if o, ok := other.(AttributeType); ok {
		return a.value == o.value
	}
	return false
}

func (a AttributeType) String() string {
	return a.value
}
