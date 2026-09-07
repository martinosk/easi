package valueobjects

import (
	"errors"
	"slices"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidSubjectType = errors.New("invalid subject type")

type SubjectType struct {
	value string
}

var subjectTypeValues = []string{
	"capability",
	"application",
	"acquired-entity",
	"vendor",
	"internal-team",
}

func NewSubjectType(value string) (SubjectType, error) {
	if slices.Contains(subjectTypeValues, value) {
		return SubjectType{value: value}, nil
	}
	return SubjectType{}, ErrInvalidSubjectType
}

func AllSubjectTypes() []SubjectType {
	all := make([]SubjectType, len(subjectTypeValues))
	for i, v := range subjectTypeValues {
		all[i] = SubjectType{value: v}
	}
	return all
}

func (s SubjectType) Value() string {
	return s.value
}

func (s SubjectType) Equals(other domain.ValueObject) bool {
	if o, ok := other.(SubjectType); ok {
		return s.value == o.value
	}
	return false
}

func (s SubjectType) String() string {
	return s.value
}
