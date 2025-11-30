package valueobjects

import (
	"easi/backend/internal/shared/domain"
	"errors"
)

var ErrInvalidImportPhase = errors.New("invalid import phase")
var ErrInvalidProgressCounts = errors.New("invalid progress counts: total must be non-negative and completed must not exceed total")

const (
	PhaseCreatingComponents         = "creating_components"
	PhaseCreatingCapabilities       = "creating_capabilities"
	PhaseCreatingRealizations       = "creating_realizations"
	PhaseCreatingComponentRelations = "creating_component_relations"
	PhaseAssigningDomains           = "assigning_domains"
)

var validPhases = map[string]bool{
	PhaseCreatingComponents:         true,
	PhaseCreatingCapabilities:       true,
	PhaseCreatingRealizations:       true,
	PhaseCreatingComponentRelations: true,
	PhaseAssigningDomains:           true,
}

type ImportProgress struct {
	phase          string
	totalItems     int
	completedItems int
}

func NewImportProgress(phase string, totalItems, completedItems int) (ImportProgress, error) {
	if !validPhases[phase] {
		return ImportProgress{}, ErrInvalidImportPhase
	}
	if totalItems < 0 || completedItems < 0 || completedItems > totalItems {
		return ImportProgress{}, ErrInvalidProgressCounts
	}
	return ImportProgress{
		phase:          phase,
		totalItems:     totalItems,
		completedItems: completedItems,
	}, nil
}

func (ip ImportProgress) Phase() string {
	return ip.phase
}

func (ip ImportProgress) TotalItems() int {
	return ip.totalItems
}

func (ip ImportProgress) CompletedItems() int {
	return ip.completedItems
}

func (ip ImportProgress) PercentComplete() int {
	if ip.totalItems == 0 {
		return 0
	}
	return (ip.completedItems * 100) / ip.totalItems
}

func (ip ImportProgress) WithIncrement(amount int) ImportProgress {
	newCompleted := ip.completedItems + amount
	if newCompleted > ip.totalItems {
		newCompleted = ip.totalItems
	}
	return ImportProgress{
		phase:          ip.phase,
		totalItems:     ip.totalItems,
		completedItems: newCompleted,
	}
}

func (ip ImportProgress) WithNextPhase(phase string, totalItems int) ImportProgress {
	return ImportProgress{
		phase:          phase,
		totalItems:     totalItems,
		completedItems: 0,
	}
}

func (ip ImportProgress) Equals(other domain.ValueObject) bool {
	if otherIP, ok := other.(ImportProgress); ok {
		return ip.phase == otherIP.phase &&
			ip.totalItems == otherIP.totalItems &&
			ip.completedItems == otherIP.completedItems
	}
	return false
}

func (ip ImportProgress) String() string {
	return ip.phase
}
