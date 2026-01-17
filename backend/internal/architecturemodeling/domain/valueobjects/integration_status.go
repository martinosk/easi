package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidIntegrationStatus = errors.New("invalid integration status: must be NOT_STARTED, IN_PROGRESS, or COMPLETED")

const (
	IntegrationStatusNotStarted = "NOT_STARTED"
	IntegrationStatusInProgress = "IN_PROGRESS"
	IntegrationStatusCompleted  = "COMPLETED"
)

type IntegrationStatus struct {
	value string
}

func NewIntegrationStatus(value string) (IntegrationStatus, error) {
	if value == "" {
		return IntegrationStatus{}, nil
	}
	upper := strings.ToUpper(strings.TrimSpace(value))
	switch upper {
	case IntegrationStatusNotStarted, IntegrationStatusInProgress, IntegrationStatusCompleted:
		return IntegrationStatus{value: upper}, nil
	default:
		return IntegrationStatus{}, ErrInvalidIntegrationStatus
	}
}

func MustNewIntegrationStatus(value string) IntegrationStatus {
	status, err := NewIntegrationStatus(value)
	if err != nil {
		panic(err)
	}
	return status
}

func (i IntegrationStatus) Value() string {
	return i.value
}

func (i IntegrationStatus) IsEmpty() bool {
	return i.value == ""
}

func (i IntegrationStatus) Equals(other domain.ValueObject) bool {
	if otherStatus, ok := other.(IntegrationStatus); ok {
		return i.value == otherStatus.value
	}
	return false
}

func (i IntegrationStatus) String() string {
	return i.value
}
