package valueobjects

import (
	"errors"
	"fmt"
)

var ErrUnknownConcern = errors.New("concern must be one of ownership, assessment, documentation, planning, structure")

type Concern struct {
	value       string
	label       string
	description string
}

var fixedConcerns = []Concern{
	{value: "ownership", label: "Ownership", description: "Every application has an owner."},
	{value: "assessment", label: "Assessment", description: "Every realisation has a current TIME grade."},
	{value: "documentation", label: "Documentation", description: "Every one-pager is complete."},
	{value: "planning", label: "Planning", description: "Eliminate grades have journeys, and journeys keep to their milestones."},
	{value: "structure", label: "Structure", description: "Every capability has a realising application and an EA owner."},
}

func NewConcern(value string) (Concern, error) {
	for _, concern := range fixedConcerns {
		if concern.value == value {
			return concern, nil
		}
	}
	return Concern{}, fmt.Errorf("%w: %q", ErrUnknownConcern, value)
}

func AllConcerns() []Concern {
	concerns := make([]Concern, len(fixedConcerns))
	copy(concerns, fixedConcerns)
	return concerns
}

func (c Concern) Value() string       { return c.value }
func (c Concern) Label() string       { return c.label }
func (c Concern) Description() string { return c.description }
func (c Concern) String() string      { return c.value }
