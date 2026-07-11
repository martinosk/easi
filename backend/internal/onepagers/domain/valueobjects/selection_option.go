package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type SelectionOption struct {
	id     OptionID
	label  OptionLabel
	active bool
}

func NewSelectionOption(id OptionID, label OptionLabel) SelectionOption {
	return SelectionOption{id: id, label: label, active: true}
}

func NewRetiredSelectionOption(id OptionID, label OptionLabel) SelectionOption {
	return SelectionOption{id: id, label: label, active: false}
}

func (s SelectionOption) ID() OptionID {
	return s.id
}

func (s SelectionOption) Label() OptionLabel {
	return s.label
}

func (s SelectionOption) IsActive() bool {
	return s.active
}

func (s SelectionOption) Retired() SelectionOption {
	s.active = false
	return s
}

func (s SelectionOption) Equals(other domain.ValueObject) bool {
	if o, ok := other.(SelectionOption); ok {
		return s.id == o.id && s.label == o.label && s.active == o.active
	}
	return false
}
