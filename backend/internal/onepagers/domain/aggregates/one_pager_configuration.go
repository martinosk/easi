package aggregates

import (
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/onepagers/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type OnePagerConfiguration struct {
	domain.AggregateRoot
	tenantID        sharedvo.TenantID
	subjectType     valueobjects.SubjectType
	displayOrder    []valueobjects.FieldRef
	builtInRequired map[string]bool
	customRequired  map[string]bool
	createdAt       valueobjects.Timestamp
	modifiedAt      valueobjects.Timestamp
	modifiedBy      valueobjects.UserEmail
}

func NewOnePagerConfiguration(
	tenantID sharedvo.TenantID,
	subjectType valueobjects.SubjectType,
	createdBy valueobjects.UserEmail,
) (*OnePagerConfiguration, error) {
	aggregate := &OnePagerConfiguration{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	entries := catalog.DefaultEntriesFor(subjectType)
	builtIns := make([]string, len(entries))
	for i, entry := range entries {
		builtIns[i] = entry.ID
	}

	event := events.NewOnePagerConfigurationCreated(events.CreateConfigurationParams{
		ID:          aggregate.ID(),
		TenantID:    tenantID.Value(),
		SubjectType: subjectType.Value(),
		BuiltIns:    builtIns,
		CreatedBy:   createdBy.Value(),
	})

	if err := aggregate.apply(event); err != nil {
		return nil, err
	}
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadOnePagerConfigurationFromHistory(eventHistory []domain.DomainEvent) (*OnePagerConfiguration, error) {
	aggregate := &OnePagerConfiguration{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	var applyErr error
	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		if applyErr != nil {
			return
		}
		applyErr = aggregate.apply(event)
	})
	if applyErr != nil {
		return nil, applyErr
	}

	return aggregate, nil
}

func (c *OnePagerConfiguration) TenantID() sharedvo.TenantID {
	return c.tenantID
}

func (c *OnePagerConfiguration) SubjectType() valueobjects.SubjectType {
	return c.subjectType
}

func (c *OnePagerConfiguration) DisplayOrder() []valueobjects.FieldRef {
	order := make([]valueobjects.FieldRef, len(c.displayOrder))
	copy(order, c.displayOrder)
	return order
}

func (c *OnePagerConfiguration) CreatedAt() valueobjects.Timestamp {
	return c.createdAt
}

func (c *OnePagerConfiguration) ModifiedAt() valueobjects.Timestamp {
	return c.modifiedAt
}

func (c *OnePagerConfiguration) ModifiedBy() valueobjects.UserEmail {
	return c.modifiedBy
}

func (c *OnePagerConfiguration) IsCustomFieldIncluded(fieldID valueobjects.FieldID) bool {
	return c.isIncluded(valueobjects.NewCustomFieldRef(fieldID))
}

func (c *OnePagerConfiguration) IsCustomFieldRequired(fieldID valueobjects.FieldID) bool {
	return c.customRequired[fieldID.Value()]
}

func (c *OnePagerConfiguration) IsBuiltInRequired(entryID string) bool {
	return c.builtInRequired[entryID]
}

func (c *OnePagerConfiguration) IncludeCustomField(fieldID valueobjects.FieldID, modifiedBy valueobjects.UserEmail) error {
	if c.IsCustomFieldIncluded(fieldID) {
		return ErrCustomFieldAlreadyIncluded
	}
	return c.applyAndRaise(events.NewCustomFieldIncluded(c.nextEventParams(modifiedBy), fieldID.Value()))
}

func (c *OnePagerConfiguration) ExcludeCustomField(fieldID valueobjects.FieldID, modifiedBy valueobjects.UserEmail) error {
	return c.onIncludedCustomField(fieldID, events.NewCustomFieldExcluded(c.nextEventParams(modifiedBy), fieldID.Value()))
}

func (c *OnePagerConfiguration) ChangeCustomFieldRequirement(
	fieldID valueobjects.FieldID,
	required bool,
	modifiedBy valueobjects.UserEmail,
) error {
	return c.onIncludedCustomField(fieldID, events.NewCustomFieldRequirementChanged(c.nextEventParams(modifiedBy), fieldID.Value(), required))
}

func (c *OnePagerConfiguration) onIncludedCustomField(fieldID valueobjects.FieldID, event domain.DomainEvent) error {
	if !c.IsCustomFieldIncluded(fieldID) {
		return ErrCustomFieldNotIncluded
	}
	return c.applyAndRaise(event)
}

func (c *OnePagerConfiguration) IncludeBuiltInField(entryID string, modifiedBy valueobjects.UserEmail) error {
	if _, found := catalog.LookupEntry(c.subjectType, entryID); !found {
		return ErrUnknownBuiltInField
	}
	if c.isBuiltInIncluded(entryID) {
		return ErrBuiltInFieldAlreadyIncluded
	}
	return c.applyAndRaise(events.NewBuiltInFieldIncluded(c.nextEventParams(modifiedBy), entryID))
}

func (c *OnePagerConfiguration) ExcludeBuiltInField(entryID string, modifiedBy valueobjects.UserEmail) error {
	if !c.isBuiltInIncluded(entryID) {
		return ErrBuiltInFieldNotIncluded
	}
	return c.applyAndRaise(events.NewBuiltInFieldExcluded(c.nextEventParams(modifiedBy), entryID))
}

func (c *OnePagerConfiguration) ChangeBuiltInFieldRequirement(entryID string, required bool, modifiedBy valueobjects.UserEmail) error {
	if _, found := catalog.LookupEntry(c.subjectType, entryID); !found {
		return ErrUnknownBuiltInField
	}
	if !c.isBuiltInIncluded(entryID) {
		return ErrBuiltInFieldNotIncluded
	}
	return c.applyAndRaise(events.NewBuiltInFieldRequirementChanged(c.nextEventParams(modifiedBy), entryID, required))
}

func (c *OnePagerConfiguration) ReorderFields(order []valueobjects.FieldRef, modifiedBy valueobjects.UserEmail) error {
	if !isPermutationOf(order, c.displayOrder) {
		return ErrInvalidDisplayOrder
	}
	return c.applyAndRaise(events.NewOnePagerFieldsReordered(c.nextEventParams(modifiedBy), refsToEventData(order)))
}

func (c *OnePagerConfiguration) nextEventParams(modifiedBy valueobjects.UserEmail) events.ModifyConfigurationParams {
	return events.ModifyConfigurationParams{
		ConfigID:   c.ID(),
		TenantID:   c.tenantID.Value(),
		Version:    c.Version() + 1,
		ModifiedBy: modifiedBy.Value(),
	}
}

func (c *OnePagerConfiguration) isBuiltInIncluded(entryID string) bool {
	ref, err := valueobjects.NewBuiltInFieldRef(entryID)
	if err != nil {
		return false
	}
	return c.isIncluded(ref)
}

func (c *OnePagerConfiguration) isIncluded(target valueobjects.FieldRef) bool {
	for _, ref := range c.displayOrder {
		if ref == target {
			return true
		}
	}
	return false
}

func isPermutationOf(proposed, current []valueobjects.FieldRef) bool {
	if len(proposed) != len(current) {
		return false
	}
	remaining := make(map[valueobjects.FieldRef]int, len(current))
	for _, ref := range current {
		remaining[ref]++
	}
	for _, ref := range proposed {
		if remaining[ref] == 0 {
			return false
		}
		remaining[ref]--
	}
	return true
}

func refsToEventData(refs []valueobjects.FieldRef) []events.FieldRefData {
	data := make([]events.FieldRefData, len(refs))
	for i, ref := range refs {
		data[i] = events.FieldRefData{Kind: string(ref.Kind()), ID: ref.RefID()}
	}
	return data
}
