package aggregates

import (
	"fmt"
	"time"

	"easi/backend/internal/metamodel/domain/events"
	"easi/backend/internal/metamodel/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

func (s *SubjectAttributeSchema) applyAndRaise(event domain.DomainEvent) error {
	if err := s.apply(event); err != nil {
		return err
	}
	s.RaiseEvent(event)
	return nil
}

func (s *SubjectAttributeSchema) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.SubjectAttributeSchemaCreated:
		return s.applyCreated(e)
	case events.SubjectAttributeDefined:
		return s.applyDefined(e)
	case events.SubjectAttributeRenamed:
		return s.applyRenamed(e)
	case events.SubjectAttributeRetired:
		return s.applyStatusChange(e.SchemaEventBase, false)
	case events.SubjectAttributeReactivated:
		return s.applyStatusChange(e.SchemaEventBase, true)
	}
	return s.applyOptionOrBoundsEvent(event)
}

func (s *SubjectAttributeSchema) applyOptionOrBoundsEvent(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.SubjectAttributeOptionAdded:
		return s.applyOptionAdded(e)
	case events.SubjectAttributeOptionRetired:
		return s.applyOptionRetired(e)
	case events.SubjectAttributeBoundsChanged:
		return s.mutateAttribute(e.SchemaEventBase, func(attribute valueobjects.SubjectAttribute) (valueobjects.SubjectAttribute, error) {
			return attribute.WithBounds(e.Min, e.Max)
		})
	}
	return nil
}

func (s *SubjectAttributeSchema) applyCreated(e events.SubjectAttributeSchemaCreated) error {
	tenantID, err := sharedvo.NewTenantID(e.TenantID)
	if err != nil {
		return fmt.Errorf("%w: tenant ID %q: %v", domain.ErrCorruptedEvent, e.TenantID, err)
	}
	subjectType, err := valueobjects.NewSubjectType(e.SubjectType)
	if err != nil {
		return fmt.Errorf("%w: subject type %q: %v", domain.ErrCorruptedEvent, e.SubjectType, err)
	}
	createdAt, err := valueobjects.NewTimestamp(e.CreatedAt)
	if err != nil {
		return fmt.Errorf("%w: created at: %v", domain.ErrCorruptedEvent, err)
	}
	createdBy, err := valueobjects.NewUserEmail(e.CreatedBy)
	if err != nil {
		return fmt.Errorf("%w: created by %q: %v", domain.ErrCorruptedEvent, e.CreatedBy, err)
	}
	s.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	s.tenantID = tenantID
	s.subjectType = subjectType
	s.attributes = nil
	s.createdAt = createdAt
	s.modifiedAt = createdAt
	s.modifiedBy = createdBy
	return nil
}

func (s *SubjectAttributeSchema) applyDefined(e events.SubjectAttributeDefined) error {
	attribute, err := attributeFromEventData(e)
	if err != nil {
		return err
	}
	s.attributes = append(s.attributes, attribute)
	return s.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func attributeFromEventData(e events.SubjectAttributeDefined) (valueobjects.SubjectAttribute, error) {
	id, err := valueobjects.NewAttributeIDFromString(e.AttributeID)
	if err != nil {
		return valueobjects.SubjectAttribute{}, fmt.Errorf("%w: attribute ID %q: %v", domain.ErrCorruptedEvent, e.AttributeID, err)
	}
	name, err := valueobjects.NewAttributeName(e.Name)
	if err != nil {
		return valueobjects.SubjectAttribute{}, fmt.Errorf("%w: attribute name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
	}
	attributeType, err := valueobjects.NewAttributeType(e.AttributeType)
	if err != nil {
		return valueobjects.SubjectAttribute{}, fmt.Errorf("%w: attribute type %q: %v", domain.ErrCorruptedEvent, e.AttributeType, err)
	}
	helpText, err := valueobjects.NewHelpText(e.HelpText)
	if err != nil {
		return valueobjects.SubjectAttribute{}, fmt.Errorf("%w: help text: %v", domain.ErrCorruptedEvent, err)
	}
	options, err := optionsFromEventData(e.Options)
	if err != nil {
		return valueobjects.SubjectAttribute{}, err
	}
	attribute, err := valueobjects.NewSubjectAttribute(valueobjects.SubjectAttributeParams{
		ID: id, Name: name, Type: attributeType, HelpText: helpText, Options: options, Min: e.Min, Max: e.Max,
	})
	if err != nil {
		return valueobjects.SubjectAttribute{}, fmt.Errorf("%w: attribute %q: %v", domain.ErrCorruptedEvent, e.AttributeID, err)
	}
	return attribute, nil
}

func optionsFromEventData(data []events.AttributeOptionData) ([]valueobjects.AttributeOption, error) {
	options := make([]valueobjects.AttributeOption, len(data))
	for i, d := range data {
		optionID, err := valueobjects.NewOptionIDFromString(d.ID)
		if err != nil {
			return nil, fmt.Errorf("%w: option ID %q: %v", domain.ErrCorruptedEvent, d.ID, err)
		}
		label, err := valueobjects.NewOptionLabel(d.Label)
		if err != nil {
			return nil, fmt.Errorf("%w: option label %q: %v", domain.ErrCorruptedEvent, d.Label, err)
		}
		if d.Active {
			options[i] = valueobjects.NewAttributeOption(optionID, label)
		} else {
			options[i] = valueobjects.NewRetiredAttributeOption(optionID, label)
		}
	}
	return options, nil
}

func (s *SubjectAttributeSchema) applyRenamed(e events.SubjectAttributeRenamed) error {
	name, err := valueobjects.NewAttributeName(e.NewName)
	if err != nil {
		return fmt.Errorf("%w: attribute name %q: %v", domain.ErrCorruptedEvent, e.NewName, err)
	}
	helpText, err := valueobjects.NewHelpText(e.NewHelpText)
	if err != nil {
		return fmt.Errorf("%w: help text: %v", domain.ErrCorruptedEvent, err)
	}
	return s.mutateAttribute(e.SchemaEventBase, func(attribute valueobjects.SubjectAttribute) (valueobjects.SubjectAttribute, error) {
		return attribute.Renamed(name, helpText), nil
	})
}

func (s *SubjectAttributeSchema) applyStatusChange(base events.SchemaEventBase, active bool) error {
	return s.mutateAttribute(base, func(attribute valueobjects.SubjectAttribute) (valueobjects.SubjectAttribute, error) {
		if active {
			return attribute.Reactivated(), nil
		}
		return attribute.Retired(), nil
	})
}

func (s *SubjectAttributeSchema) applyOptionAdded(e events.SubjectAttributeOptionAdded) error {
	optionID, err := valueobjects.NewOptionIDFromString(e.OptionID)
	if err != nil {
		return fmt.Errorf("%w: option ID %q: %v", domain.ErrCorruptedEvent, e.OptionID, err)
	}
	label, err := valueobjects.NewOptionLabel(e.Label)
	if err != nil {
		return fmt.Errorf("%w: option label %q: %v", domain.ErrCorruptedEvent, e.Label, err)
	}
	return s.mutateAttribute(e.SchemaEventBase, func(attribute valueobjects.SubjectAttribute) (valueobjects.SubjectAttribute, error) {
		return attribute.WithAddedOption(valueobjects.NewAttributeOption(optionID, label))
	})
}

func (s *SubjectAttributeSchema) applyOptionRetired(e events.SubjectAttributeOptionRetired) error {
	optionID, err := valueobjects.NewOptionIDFromString(e.OptionID)
	if err != nil {
		return fmt.Errorf("%w: option ID %q: %v", domain.ErrCorruptedEvent, e.OptionID, err)
	}
	return s.mutateAttribute(e.SchemaEventBase, func(attribute valueobjects.SubjectAttribute) (valueobjects.SubjectAttribute, error) {
		return attribute.WithRetiredOption(optionID)
	})
}

func (s *SubjectAttributeSchema) mutateAttribute(
	base events.SchemaEventBase,
	mutate func(valueobjects.SubjectAttribute) (valueobjects.SubjectAttribute, error),
) error {
	index := s.findAttributeIndex(base.AttributeID)
	if index < 0 {
		return fmt.Errorf("%w: attribute %q not found", domain.ErrCorruptedEvent, base.AttributeID)
	}
	mutated, err := mutate(s.attributes[index])
	if err != nil {
		return fmt.Errorf("%w: attribute %q: %v", domain.ErrCorruptedEvent, base.AttributeID, err)
	}
	s.attributes[index] = mutated
	return s.applyModificationMetadata(base.ModifiedAt, base.ModifiedBy)
}

func (s *SubjectAttributeSchema) applyModificationMetadata(modifiedAtRaw time.Time, modifiedByRaw string) error {
	modifiedAt, err := valueobjects.NewTimestamp(modifiedAtRaw)
	if err != nil {
		return fmt.Errorf("%w: modified at: %v", domain.ErrCorruptedEvent, err)
	}
	modifiedBy, err := valueobjects.NewUserEmail(modifiedByRaw)
	if err != nil {
		return fmt.Errorf("%w: modified by %q: %v", domain.ErrCorruptedEvent, modifiedByRaw, err)
	}
	s.modifiedAt = modifiedAt
	s.modifiedBy = modifiedBy
	return nil
}
