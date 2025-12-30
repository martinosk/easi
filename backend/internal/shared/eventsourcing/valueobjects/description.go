package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const MaxDescriptionLength = 1000

var ErrDescriptionTooLong = errors.New("description exceeds maximum length of 1000 characters")

type Description struct {
	value string
}

func NewDescription(value string) (Description, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > MaxDescriptionLength {
		return Description{}, ErrDescriptionTooLong
	}
	return Description{value: trimmed}, nil
}

func MustNewDescription(value string) Description {
	desc, err := NewDescription(value)
	if err != nil {
		panic(err)
	}
	return desc
}

func (d Description) Value() string {
	return d.value
}

func (d Description) IsEmpty() bool {
	return d.value == ""
}

func (d Description) Equals(other domain.ValueObject) bool {
	if otherDesc, ok := other.(Description); ok {
		return d.value == otherDesc.value
	}
	return false
}

func (d Description) String() string {
	return d.value
}
