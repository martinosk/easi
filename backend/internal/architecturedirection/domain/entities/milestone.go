package entities

import (
	"errors"
	"strings"

	"easi/backend/internal/architecturedirection/domain/valueobjects"
)

const MaxMilestoneLabelLength = 200

var (
	ErrMilestoneLabelRequired = errors.New("milestone label is required")
	ErrMilestoneLabelTooLong  = errors.New("milestone label exceeds maximum length of 200 characters")
)

type Milestone struct {
	id           string
	label        string
	targetPeriod *valueobjects.TargetPeriod
	status       valueobjects.MilestoneStatus
}

func NewMilestone(id, label string, targetPeriod *valueobjects.TargetPeriod, status valueobjects.MilestoneStatus) (Milestone, error) {
	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		return Milestone{}, ErrMilestoneLabelRequired
	}
	if len(trimmed) > MaxMilestoneLabelLength {
		return Milestone{}, ErrMilestoneLabelTooLong
	}
	return Milestone{id: id, label: trimmed, targetPeriod: targetPeriod, status: status}, nil
}

func (m Milestone) ID() string                               { return m.id }
func (m Milestone) Label() string                            { return m.label }
func (m Milestone) TargetPeriod() *valueobjects.TargetPeriod { return m.targetPeriod }
func (m Milestone) Status() valueobjects.MilestoneStatus     { return m.status }
