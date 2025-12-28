package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/google/uuid"
)

var (
	ErrAssignmentIDMissingPrefix = errors.New("assignment ID must start with 'assign-' prefix")
)

type AssignmentID struct {
	value string
}

func NewAssignmentID() AssignmentID {
	return AssignmentID{value: "assign-" + uuid.New().String()}
}

func NewAssignmentIDFromString(value string) (AssignmentID, error) {
	if value == "" {
		return AssignmentID{}, domain.ErrEmptyValue
	}

	if !strings.HasPrefix(value, "assign-") {
		return AssignmentID{}, ErrAssignmentIDMissingPrefix
	}

	guidPart := strings.TrimPrefix(value, "assign-")
	if _, err := uuid.Parse(guidPart); err != nil {
		return AssignmentID{}, domain.ErrInvalidValue
	}

	return AssignmentID{value: value}, nil
}

func (a AssignmentID) Value() string {
	return a.value
}

func (a AssignmentID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(AssignmentID); ok {
		return a.value == otherID.value
	}
	return false
}

func (a AssignmentID) String() string {
	return a.value
}
