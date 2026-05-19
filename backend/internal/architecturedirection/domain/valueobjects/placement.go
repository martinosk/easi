package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

const MaxResultingNameLength = 200

var ErrResultingNameTooLong = errors.New("resulting name cannot exceed 200 characters")

type Placement struct {
	targetBusinessDomainID string
	resultingName          string
}

func NewPlacement(targetBusinessDomainID, resultingName string) (Placement, error) {
	if _, err := sharedvo.NewUUIDValueFromString(targetBusinessDomainID); err != nil {
		return Placement{}, err
	}
	name := strings.TrimSpace(resultingName)
	if len(name) > MaxResultingNameLength {
		return Placement{}, ErrResultingNameTooLong
	}
	return Placement{targetBusinessDomainID: targetBusinessDomainID, resultingName: name}, nil
}

func (p Placement) TargetBusinessDomainID() string { return p.targetBusinessDomainID }
func (p Placement) ResultingName() string          { return p.resultingName }

func (p Placement) Equals(other domain.ValueObject) bool {
	if o, ok := other.(Placement); ok {
		return p.targetBusinessDomainID == o.targetBusinessDomainID && p.resultingName == o.resultingName
	}
	return false
}
