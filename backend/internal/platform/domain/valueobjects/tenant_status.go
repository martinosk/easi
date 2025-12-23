package valueobjects

import (
	"easi/backend/internal/shared/eventsourcing"
	"errors"
)

var (
	ErrInvalidTenantStatus = errors.New("invalid tenant status: must be active, suspended, or archived")
)

type TenantStatus struct {
	value string
}

var (
	TenantStatusActive    = TenantStatus{value: "active"}
	TenantStatusSuspended = TenantStatus{value: "suspended"}
	TenantStatusArchived  = TenantStatus{value: "archived"}
)

func NewTenantStatus(value string) (TenantStatus, error) {
	switch value {
	case "active":
		return TenantStatusActive, nil
	case "suspended":
		return TenantStatusSuspended, nil
	case "archived":
		return TenantStatusArchived, nil
	default:
		return TenantStatus{}, ErrInvalidTenantStatus
	}
}

func (s TenantStatus) Value() string {
	return s.value
}

func (s TenantStatus) String() string {
	return s.value
}

func (s TenantStatus) IsActive() bool {
	return s.value == "active"
}

func (s TenantStatus) Equals(other domain.ValueObject) bool {
	if otherStatus, ok := other.(TenantStatus); ok {
		return s.value == otherStatus.value
	}
	return false
}
