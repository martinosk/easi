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
	case events.CustomFieldIncluded:
		return c.applyCustomFieldInclusion(e.FieldID, e.ConfigurationEventBase, true)
	case events.CustomFieldExcluded:
		return c.applyCustomFieldInclusion(e.FieldID, e.ConfigurationEventBase, false)
	case events.CustomFieldRequirementChanged:
		return c.applyCustomFieldRequirementChanged(e)
	}
	return c.applyLegacySchemaEvent(event)
}

func (c *OnePagerConfiguration) applyLegacySchemaEvent(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.CustomFieldDefined:
		return c.applyLegacyCustomFieldDefined(e)
	case events.CustomFieldRetired:
		return c.applyCustomFieldInclusion(e.FieldID, e.ConfigurationEventBase, false)
	case events.CustomFieldReactivated:
		return c.applyCustomFieldInclusion(e.FieldID, e.ConfigurationEventBase, true)
	case events.CustomFieldRenamed:
		return c.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
	case events.SelectionOptionAdded:
		return c.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
	case events.SelectionOptionRetired:
		return c.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
	case events.NumberFieldBoundsChanged:
		return c.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
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
	c.displayOrder = displayOrder
	c.builtInRequired = map[string]bool{}
	c.customRequired = map[string]bool{}
	c.createdAt = createdAt
	c.modifiedAt = createdAt
	c.modifiedBy = createdBy
	return nil
}

func (c *OnePagerConfiguration) applyLegacyCustomFieldDefined(e events.CustomFieldDefined) error {
	c.customRequired[e.FieldID] = e.Required
	return c.applyCustomFieldInclusion(e.FieldID, e.ConfigurationEventBase, true)
}

func (c *OnePagerConfiguration) applyCustomFieldInclusion(rawFieldID string, base events.ConfigurationEventBase, included bool) error {
	fieldID, err := valueobjects.NewFieldIDFromString(rawFieldID)
	if err != nil {
		return fmt.Errorf("%w: field ID %q: %v", domain.ErrCorruptedEvent, rawFieldID, err)
	}
	ref := valueobjects.NewCustomFieldRef(fieldID)
	c.removeFromDisplayOrder(ref)
	if included {
		c.displayOrder = append(c.displayOrder, ref)
	}
	return c.applyModificationMetadata(base.ModifiedAt, base.ModifiedBy)
}

func (c *OnePagerConfiguration) applyCustomFieldRequirementChanged(e events.CustomFieldRequirementChanged) error {
	c.customRequired[e.FieldID] = e.Required
	return c.applyModificationMetadata(e.ModifiedAt, e.ModifiedBy)
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
