package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
)

var ErrInvalidImportStatus = errors.New("invalid import status: must be one of 'pending', 'importing', 'completed', 'failed'")

const (
	statusPending   = "pending"
	statusImporting = "importing"
	statusCompleted = "completed"
	statusFailed    = "failed"
)

var validStatuses = map[string]bool{
	statusPending:   true,
	statusImporting: true,
	statusCompleted: true,
	statusFailed:    true,
}

type ImportStatus struct {
	value string
}

func NewImportStatus(value string) (ImportStatus, error) {
	if !validStatuses[value] {
		return ImportStatus{}, ErrInvalidImportStatus
	}
	return ImportStatus{value: value}, nil
}

func ImportStatusPending() ImportStatus {
	return ImportStatus{value: statusPending}
}

func ImportStatusImporting() ImportStatus {
	return ImportStatus{value: statusImporting}
}

func ImportStatusCompleted() ImportStatus {
	return ImportStatus{value: statusCompleted}
}

func ImportStatusFailed() ImportStatus {
	return ImportStatus{value: statusFailed}
}

func (is ImportStatus) Value() string {
	return is.value
}

func (is ImportStatus) IsPending() bool {
	return is.value == statusPending
}

func (is ImportStatus) IsImporting() bool {
	return is.value == statusImporting
}

func (is ImportStatus) IsCompleted() bool {
	return is.value == statusCompleted
}

func (is ImportStatus) IsFailed() bool {
	return is.value == statusFailed
}

func (is ImportStatus) CanTransitionTo(target ImportStatus) bool {
	switch is.value {
	case statusPending:
		return target.value == statusImporting
	case statusImporting:
		return target.value == statusCompleted || target.value == statusFailed
	default:
		return false
	}
}

func (is ImportStatus) Equals(other domain.ValueObject) bool {
	if otherIS, ok := other.(ImportStatus); ok {
		return is.value == otherIS.value
	}
	return false
}

func (is ImportStatus) String() string {
	return is.value
}
