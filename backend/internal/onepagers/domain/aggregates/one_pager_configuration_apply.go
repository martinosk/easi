package aggregates

import (
	"fmt"
	"time"

	"easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/onepagers/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

func (c *OnePagerConfiguration) applyAndRaise(event domain.DomainEvent) error {
	if err := c.apply(event); err != nil {
		return err
	}
	c.RaiseEvent(event)
	return nil
}

func (c *OnePagerConfiguration) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.OnePagerConfigurationCreated:
		return c.applyCreated(e)
	case events.BuiltInFieldIncluded:
		return c.applyBuiltInFieldIncluded(e)
	case events.BuiltInFieldExcluded:
		return c.applyBuiltInFieldExcluded(e)
	case events.BuiltInFieldRequirementChanged:
		return c.applyBuiltInFieldRequirementChanged(e)
	case events.OnePagerFieldsReordered:
		return c.applyFieldsReordered(e)
	}
	return c.applyCustomFieldEvent(event)
}

func (c *OnePagerConfiguration) applyCustomFieldEvent(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.CustomFieldDefined:
		return c.applyCustomFieldDefined(e)
	case events.CustomFieldRenamed:
		return c.applyCustomFieldRenamed(e)
	case events.CustomFieldRequirementChanged:
		return c.applyCustomFieldRequirementChanged(e)
	case events.CustomFieldRetired:
		return c.applyCustomFieldRetired(e)
	case events.CustomFieldReactivated:
		return c.applyCustomFieldReactivated(e)
	}
	return c.applyCustomFieldAttributeEvent(event)
}

func (c *OnePagerConfiguration) applyCustomFieldAttributeEvent(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.SelectionOptionAdded:
		return c.applySelectionOptionAdded(e)
	case events.SelectionOptionRetired:
		return c.applySelectionOptionRetired(e)
	case events.NumberFieldBoundsChanged:
		return c.applyNumberFieldBoundsChanged(e)
	}
	return nil
}

func (c *OnePagerConfiguration) applyCreated(e events.OnePagerConfigurationCreated) error {
	tenantID, err := sharedvo.NewTenantID(e.TenantID)
	if err != nil {
		return fmt.Errorf("%w: tenant ID %q: %v", domain.ErrCorruptedEvent, e.TenantID, err)
	}
	subjectType, err := valueobjects.NewSubjectType(e.SubjectType)
	if err != nil {
		return fmt.Errorf("%w: subject type %q: %v", domain.ErrCorruptedEvent, e.SubjectType, err)
	}
	displayOrder := make([]valueobjects.FieldRef, len(e.BuiltIns))
	for i, entryID := range e.BuiltIns {
		ref, err := valueobjects.NewBuiltInFieldRef(entryID)
		if err != nil {
			return fmt.Errorf("%w: built-in entry %q: %v", domain.ErrCorruptedEvent, entryID, err)
		}
		displayOrder[i] = ref
	}
	createdAt, err := valueobjects.NewTimestamp(e.CreatedAt)
	if err != nil {
		return fmt.Errorf("%w: created at: %v", domain.ErrCorruptedEvent, err)
	}
	createdBy, err := valueobjects.NewUserEmail(e.CreatedBy)
	if err != nil {
		return fmt.Errorf("%w: created by %q: %v", domain.ErrCorruptedEvent, e.CreatedBy, err)
	}

	c.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	c.tenantID = tenantID
	c.subjectType = subjectType
	c.customFields = nil
	c.displayOrder = displayOrder
	c.builtInRequired = map[string]bool{}
	c.createdAt = createdAt
	c.modifiedAt = createdAt
	c.modifiedBy = createdBy
	return nil
}

func (c *OnePagerConfiguration) applyCustomFieldDefined(e events.CustomFieldDefined) error {
	field, err := customFieldFromEventData(e)
	if err != nil {
		return err
	}
	c.customFields = append(c.customFields, field)
	c.displayOrder = append(c.displayOrder, valueobjects.NewCustomFieldRef(field.ID()))
	return c.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func customFieldFromEventData(e events.CustomFieldDefined) (valueobjects.CustomField, error) {
	fieldID, err := valueobjects.NewFieldIDFromString(e.FieldID)
	if err != nil {
		return valueobjects.CustomField{}, fmt.Errorf("%w: field ID %q: %v", domain.ErrCorruptedEvent, e.FieldID, err)
	}
	name, err := valueobjects.NewFieldName(e.Name)
	if err != nil {
		return valueobjects.CustomField{}, fmt.Errorf("%w: field name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
	}
	fieldType, err := valueobjects.NewFieldType(e.FieldType)
	if err != nil {
		return valueobjects.CustomField{}, fmt.Errorf("%w: field type %q: %v", domain.ErrCorruptedEvent, e.FieldType, err)
	}
	helpText, err := valueobjects.NewHelpText(e.HelpText)
	if err != nil {
		return valueobjects.CustomField{}, fmt.Errorf("%w: help text: %v", domain.ErrCorruptedEvent, err)
	}
	options, err := optionsFromEventData(e.Options)
	if err != nil {
		return valueobjects.CustomField{}, err
	}
	field, err := valueobjects.NewCustomField(valueobjects.CustomFieldParams{
		ID:       fieldID,
		Name:     name,
		Type:     fieldType,
		Required: e.Required,
		HelpText: helpText,
		Options:  options,
	})
	if err != nil {
		return valueobjects.CustomField{}, fmt.Errorf("%w: custom field: %v", domain.ErrCorruptedEvent, err)
	}
	return field, nil
}

func optionsFromEventData(data []events.SelectionOptionData) ([]valueobjects.SelectionOption, error) {
	options := make([]valueobjects.SelectionOption, len(data))
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
			options[i] = valueobjects.NewSelectionOption(optionID, label)
		} else {
			options[i] = valueobjects.NewRetiredSelectionOption(optionID, label)
		}
	}
	return options, nil
}

func (c *OnePagerConfiguration) applyCustomFieldRenamed(e events.CustomFieldRenamed) error {
	name, err := valueobjects.NewFieldName(e.NewName)
	if err != nil {
		return fmt.Errorf("%w: field name %q: %v", domain.ErrCorruptedEvent, e.NewName, err)
	}
	helpText, err := valueobjects.NewHelpText(e.NewHelpText)
	if err != nil {
		return fmt.Errorf("%w: help text: %v", domain.ErrCorruptedEvent, err)
	}
	return c.mutateField(e.FieldID, e.ModifiedAt, e.ModifiedBy, func(field valueobjects.CustomField) (valueobjects.CustomField, error) {
		return field.Renamed(name, helpText), nil
	})
}

func (c *OnePagerConfiguration) applyCustomFieldRequirementChanged(e events.CustomFieldRequirementChanged) error {
	return c.mutateField(e.FieldID, e.ModifiedAt, e.ModifiedBy, func(field valueobjects.CustomField) (valueobjects.CustomField, error) {
		return field.WithRequirement(e.Required), nil
	})
}

func (c *OnePagerConfiguration) applyCustomFieldRetired(e events.CustomFieldRetired) error {
	return c.applyFieldStatusChange(e.FieldID, e.ConfigurationEventBase, false)
}

func (c *OnePagerConfiguration) applyCustomFieldReactivated(e events.CustomFieldReactivated) error {
	return c.applyFieldStatusChange(e.FieldID, e.ConfigurationEventBase, true)
}

func (c *OnePagerConfiguration) applyFieldStatusChange(rawFieldID string, base events.ConfigurationEventBase, active bool) error {
	fieldID, err := valueobjects.NewFieldIDFromString(rawFieldID)
	if err != nil {
		return fmt.Errorf("%w: field ID %q: %v", domain.ErrCorruptedEvent, rawFieldID, err)
	}
	ref := valueobjects.NewCustomFieldRef(fieldID)
	if active {
		c.displayOrder = append(c.displayOrder, ref)
	} else {
		c.removeFromDisplayOrder(ref)
	}
	return c.mutateField(rawFieldID, base.ModifiedAt, base.ModifiedBy, func(field valueobjects.CustomField) (valueobjects.CustomField, error) {
		if active {
			return field.Reactivated(), nil
		}
		return field.Retired(), nil
	})
}

func (c *OnePagerConfiguration) applyBuiltInFieldIncluded(e events.BuiltInFieldIncluded) error {
	ref, err := valueobjects.NewBuiltInFieldRef(e.EntryID)
	if err != nil {
		return fmt.Errorf("%w: built-in entry %q: %v", domain.ErrCorruptedEvent, e.EntryID, err)
	}
	c.displayOrder = append(c.displayOrder, ref)
	return c.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func (c *OnePagerConfiguration) applyBuiltInFieldExcluded(e events.BuiltInFieldExcluded) error {
	ref, err := valueobjects.NewBuiltInFieldRef(e.EntryID)
	if err != nil {
		return fmt.Errorf("%w: built-in entry %q: %v", domain.ErrCorruptedEvent, e.EntryID, err)
	}
	c.removeFromDisplayOrder(ref)
	return c.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func (c *OnePagerConfiguration) applyBuiltInFieldRequirementChanged(e events.BuiltInFieldRequirementChanged) error {
	if c.builtInRequired == nil {
		c.builtInRequired = map[string]bool{}
	}
	c.builtInRequired[e.EntryID] = e.Required
	return c.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func (c *OnePagerConfiguration) applyFieldsReordered(e events.OnePagerFieldsReordered) error {
	order := make([]valueobjects.FieldRef, len(e.Order))
	for i, d := range e.Order {
		ref, err := valueobjects.NewFieldRef(d.Kind, d.ID)
		if err != nil {
			return fmt.Errorf("%w: field ref %q/%q: %v", domain.ErrCorruptedEvent, d.Kind, d.ID, err)
		}
		order[i] = ref
	}
	c.displayOrder = order
	return c.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
}

func (c *OnePagerConfiguration) applySelectionOptionAdded(e events.SelectionOptionAdded) error {
	optionID, err := valueobjects.NewOptionIDFromString(e.OptionID)
	if err != nil {
		return fmt.Errorf("%w: option ID %q: %v", domain.ErrCorruptedEvent, e.OptionID, err)
	}
	label, err := valueobjects.NewOptionLabel(e.Label)
	if err != nil {
		return fmt.Errorf("%w: option label %q: %v", domain.ErrCorruptedEvent, e.Label, err)
	}
	return c.mutateField(e.FieldID, e.ModifiedAt, e.ModifiedBy, func(field valueobjects.CustomField) (valueobjects.CustomField, error) {
		return field.WithAddedOption(valueobjects.NewSelectionOption(optionID, label))
	})
}

func (c *OnePagerConfiguration) applySelectionOptionRetired(e events.SelectionOptionRetired) error {
	optionID, err := valueobjects.NewOptionIDFromString(e.OptionID)
	if err != nil {
		return fmt.Errorf("%w: option ID %q: %v", domain.ErrCorruptedEvent, e.OptionID, err)
	}
	return c.mutateField(e.FieldID, e.ModifiedAt, e.ModifiedBy, func(field valueobjects.CustomField) (valueobjects.CustomField, error) {
		return field.WithRetiredOption(optionID)
	})
}

func (c *OnePagerConfiguration) applyNumberFieldBoundsChanged(e events.NumberFieldBoundsChanged) error {
	return c.mutateField(e.FieldID, e.ModifiedAt, e.ModifiedBy, func(field valueobjects.CustomField) (valueobjects.CustomField, error) {
		return field.WithBounds(e.Min, e.Max)
	})
}

func (c *OnePagerConfiguration) mutateField(
	fieldID string,
	modifiedAt time.Time,
	modifiedBy string,
	mutate func(valueobjects.CustomField) (valueobjects.CustomField, error),
) error {
	index := c.findFieldIndex(fieldID)
	if index < 0 {
		return fmt.Errorf("%w: custom field %q not found", domain.ErrCorruptedEvent, fieldID)
	}
	mutated, err := mutate(c.customFields[index])
	if err != nil {
		return fmt.Errorf("%w: custom field %q: %v", domain.ErrCorruptedEvent, fieldID, err)
	}
	c.customFields[index] = mutated
	return c.applyModificationMetadata(modifiedAt, modifiedBy)
}

func (c *OnePagerConfiguration) removeFromDisplayOrder(target valueobjects.FieldRef) {
	order := make([]valueobjects.FieldRef, 0, len(c.displayOrder))
	for _, ref := range c.displayOrder {
		if ref != target {
			order = append(order, ref)
		}
	}
	c.displayOrder = order
}

func (c *OnePagerConfiguration) applyModificationMetadata(modifiedAtRaw time.Time, modifiedByRaw string) error {
	modifiedAt, err := valueobjects.NewTimestamp(modifiedAtRaw)
	if err != nil {
		return fmt.Errorf("%w: modified at: %v", domain.ErrCorruptedEvent, err)
	}
	modifiedBy, err := valueobjects.NewUserEmail(modifiedByRaw)
	if err != nil {
		return fmt.Errorf("%w: modified by %q: %v", domain.ErrCorruptedEvent, modifiedByRaw, err)
	}
	c.modifiedAt = modifiedAt
	c.modifiedBy = modifiedBy
	return nil
}
