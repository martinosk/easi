package valueobjects

import (
	"errors"
	"strings"

	"easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidOwnershipModel = errors.New("invalid ownership model: must be TribeOwned, TeamOwned, Shared, or EnterpriseService")
)

type OwnershipModel string

const (
	OwnershipTribeOwned        OwnershipModel = "TribeOwned"
	OwnershipTeamOwned         OwnershipModel = "TeamOwned"
	OwnershipShared            OwnershipModel = "Shared"
	OwnershipEnterpriseService OwnershipModel = "EnterpriseService"
)

func NewOwnershipModel(value string) (OwnershipModel, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", nil
	}

	switch OwnershipModel(normalized) {
	case OwnershipTribeOwned, OwnershipTeamOwned, OwnershipShared, OwnershipEnterpriseService:
		return OwnershipModel(normalized), nil
	default:
		return "", ErrInvalidOwnershipModel
	}
}

func (o OwnershipModel) Value() string {
	return string(o)
}

func (o OwnershipModel) Equals(other domain.ValueObject) bool {
	if otherModel, ok := other.(OwnershipModel); ok {
		return o == otherModel
	}
	return false
}

func (o OwnershipModel) String() string {
	return string(o)
}

func (o OwnershipModel) IsEmpty() bool {
	return o == ""
}
