package valueobjects

import (
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

type Description struct {
	value string
}

func NewDescription(value string) Description {
	return Description{value: strings.TrimSpace(value)}
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
