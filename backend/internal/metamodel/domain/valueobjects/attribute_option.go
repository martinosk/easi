package valueobjects

import domain "easi/backend/internal/shared/eventsourcing"

type AttributeOption struct {
	id     OptionID
	label  OptionLabel
	active bool
}

func NewAttributeOption(id OptionID, label OptionLabel) AttributeOption {
	return AttributeOption{id: id, label: label, active: true}
}

func NewRetiredAttributeOption(id OptionID, label OptionLabel) AttributeOption {
	return AttributeOption{id: id, label: label, active: false}
}

func (o AttributeOption) ID() OptionID {
	return o.id
}

func (o AttributeOption) Label() OptionLabel {
	return o.label
}

func (o AttributeOption) IsActive() bool {
	return o.active
}

func (o AttributeOption) Retired() AttributeOption {
	o.active = false
	return o
}

func (o AttributeOption) Equals(other domain.ValueObject) bool {
	if other2, ok := other.(AttributeOption); ok {
		return o.id == other2.id && o.label == other2.label && o.active == other2.active
	}
	return false
}
