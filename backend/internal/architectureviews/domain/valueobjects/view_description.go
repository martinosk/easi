package valueobjects

import (
	"strings"

	"easi/backend/internal/shared/domain"
)

const MaxDescriptionLength = 500

type ViewDescription struct {
	value string
}

func NewViewDescription(value string) ViewDescription {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > MaxDescriptionLength {
		trimmed = trimmed[:MaxDescriptionLength]
	}
	return ViewDescription{value: trimmed}
}

func (d ViewDescription) Value() string {
	return d.value
}

func (d ViewDescription) IsEmpty() bool {
	return d.value == ""
}

func (d ViewDescription) Equals(other domain.ValueObject) bool {
	if otherDesc, ok := other.(ViewDescription); ok {
		return d.value == otherDesc.value
	}
	return false
}

func (d ViewDescription) String() string {
	return d.value
}