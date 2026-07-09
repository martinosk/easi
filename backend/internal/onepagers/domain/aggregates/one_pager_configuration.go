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
	tenantID     sharedvo.TenantID
	subjectType  valueobjects.SubjectType
	customFields []valueobjects.CustomField
	displayOrder []valueobjects.FieldRef
	createdAt    valueobjects.Timestamp
	modifiedAt   valueobjects.Timestamp
	modifiedBy   valueobjects.UserEmail
}

func NewOnePagerConfiguration(
	tenantID sharedvo.TenantID,
	subjectType valueobjects.SubjectType,
	createdBy valueobjects.UserEmail,
) (*OnePagerConfiguration, error) {
	aggregate := &OnePagerConfiguration{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	entries := catalog.EntriesFor(subjectType)
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

func (c *OnePagerConfiguration) CustomFields() []valueobjects.CustomField {
	fields := make([]valueobjects.CustomField, len(c.customFields))
	copy(fields, c.customFields)
	return fields
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

func (c *OnePagerConfiguration) CustomFieldByID(fieldID valueobjects.FieldID) (valueobjects.CustomField, bool) {
	index := c.findFieldIndex(fieldID.Value())
	if index < 0 {
		return valueobjects.CustomField{}, false
	}
	return c.customFields[index], true
}

type DefineCustomFieldParams struct {
	Name         valueobjects.FieldName
	Type         valueobjects.FieldType
	Required     bool
	HelpText     valueobjects.HelpText
	OptionLabels []valueobjects.OptionLabel
}

func (c *OnePagerConfiguration) DefineCustomField(
	params DefineCustomFieldParams,
	modifiedBy valueobjects.UserEmail,
) (valueobjects.FieldID, error) {
	fieldID := valueobjects.NewFieldID()
	options := make([]valueobjects.SelectionOption, len(params.OptionLabels))
	for i, label := range params.OptionLabels {
		options[i] = valueobjects.NewSelectionOption(valueobjects.NewOptionID(), label)
	}

	if _, err := valueobjects.NewCustomField(valueobjects.CustomFieldParams{
		ID:       fieldID,
		Name:     params.Name,
		Type:     params.Type,
		Required: params.Required,
		HelpText: params.HelpText,
		Options:  options,
	}); err != nil {
		return valueobjects.FieldID{}, err
	}

	if c.activeNameExists(params.Name, "") {
		return valueobjects.FieldID{}, ErrDuplicateFieldName
	}

	event := events.NewCustomFieldDefined(c.nextEventParams(modifiedBy), events.CustomFieldData{
		FieldID:   fieldID.Value(),
		Name:      params.Name.Value(),
		FieldType: params.Type.Value(),
		Required:  params.Required,
		HelpText:  params.HelpText.Value(),
		Options:   optionsToEventData(options),
	})
	if err := c.applyAndRaise(event); err != nil {
		return valueobjects.FieldID{}, err
	}
	return fieldID, nil
}

type RenameCustomFieldParams struct {
	FieldID       valueobjects.FieldID
	Name          valueobjects.FieldName
	HelpText      valueobjects.HelpText
	RequestedType string
}

func (c *OnePagerConfiguration) RenameCustomField(
	params RenameCustomFieldParams,
	modifiedBy valueobjects.UserEmail,
) error {
	field, err := c.activeField(params.FieldID)
	if err != nil {
		return err
	}
	if params.RequestedType != "" && params.RequestedType != field.Type().Value() {
		return ErrFieldTypeImmutable
	}
	if c.activeNameExists(params.Name, params.FieldID.Value()) {
		return ErrDuplicateFieldName
	}

	event := events.NewCustomFieldRenamed(c.nextEventParams(modifiedBy), events.FieldRenameData{
		FieldID:     params.FieldID.Value(),
		NewName:     params.Name.Value(),
		NewHelpText: params.HelpText.Value(),
	})
	return c.applyAndRaise(event)
}

func (c *OnePagerConfiguration) ChangeCustomFieldRequirement(
	fieldID valueobjects.FieldID,
	required bool,
	modifiedBy valueobjects.UserEmail,
) error {
	if _, err := c.activeField(fieldID); err != nil {
		return err
	}

	event := events.NewCustomFieldRequirementChanged(c.nextEventParams(modifiedBy), fieldID.Value(), required)
	return c.applyAndRaise(event)
}

func (c *OnePagerConfiguration) RetireCustomField(fieldID valueobjects.FieldID, modifiedBy valueobjects.UserEmail) error {
	index := c.findFieldIndex(fieldID.Value())
	if index < 0 {
		return ErrFieldNotFound
	}
	if !c.customFields[index].IsActive() {
		return ErrFieldAlreadyRetired
	}

	event := events.NewCustomFieldRetired(c.nextEventParams(modifiedBy), fieldID.Value())
	return c.applyAndRaise(event)
}

func (c *OnePagerConfiguration) ReactivateCustomField(fieldID valueobjects.FieldID, modifiedBy valueobjects.UserEmail) error {
	index := c.findFieldIndex(fieldID.Value())
	if index < 0 {
		return ErrFieldNotFound
	}
	field := c.customFields[index]
	if field.IsActive() {
		return ErrFieldAlreadyActive
	}
	if c.activeNameExists(field.Name(), fieldID.Value()) {
		return ErrDuplicateFieldName
	}

	event := events.NewCustomFieldReactivated(c.nextEventParams(modifiedBy), fieldID.Value())
	return c.applyAndRaise(event)
}

func (c *OnePagerConfiguration) IncludeBuiltInField(entryID string, modifiedBy valueobjects.UserEmail) error {
	entry, found := catalog.LookupEntry(c.subjectType, entryID)
	if !found {
		return ErrUnknownBuiltInField
	}
	if c.isBuiltInIncluded(entryID) {
		return ErrBuiltInFieldAlreadyIncluded
	}
	if label, err := valueobjects.NewFieldName(entry.Label); err == nil && c.activeNameExists(label, "") {
		return ErrDuplicateFieldName
	}

	event := events.NewBuiltInFieldIncluded(c.nextEventParams(modifiedBy), entryID)
	return c.applyAndRaise(event)
}

func (c *OnePagerConfiguration) ExcludeBuiltInField(entryID string, modifiedBy valueobjects.UserEmail) error {
	if !c.isBuiltInIncluded(entryID) {
		return ErrBuiltInFieldNotIncluded
	}

	event := events.NewBuiltInFieldExcluded(c.nextEventParams(modifiedBy), entryID)
	return c.applyAndRaise(event)
}

func (c *OnePagerConfiguration) ReorderFields(order []valueobjects.FieldRef, modifiedBy valueobjects.UserEmail) error {
	if !isPermutationOf(order, c.displayOrder) {
		return ErrInvalidDisplayOrder
	}

	event := events.NewOnePagerFieldsReordered(c.nextEventParams(modifiedBy), refsToEventData(order))
	return c.applyAndRaise(event)
}

func (c *OnePagerConfiguration) AddSelectionOption(
	fieldID valueobjects.FieldID,
	label valueobjects.OptionLabel,
	modifiedBy valueobjects.UserEmail,
) (valueobjects.OptionID, error) {
	field, err := c.activeField(fieldID)
	if err != nil {
		return valueobjects.OptionID{}, err
	}

	optionID := valueobjects.NewOptionID()
	if _, err := field.WithAddedOption(valueobjects.NewSelectionOption(optionID, label)); err != nil {
		return valueobjects.OptionID{}, err
	}

	event := events.NewSelectionOptionAdded(c.nextEventParams(modifiedBy), fieldID.Value(), optionID.Value(), label.Value())
	if err := c.applyAndRaise(event); err != nil {
		return valueobjects.OptionID{}, err
	}
	return optionID, nil
}

func (c *OnePagerConfiguration) RetireSelectionOption(
	fieldID valueobjects.FieldID,
	optionID valueobjects.OptionID,
	modifiedBy valueobjects.UserEmail,
) error {
	field, err := c.activeField(fieldID)
	if err != nil {
		return err
	}
	if _, err := field.WithRetiredOption(optionID); err != nil {
		return err
	}

	event := events.NewSelectionOptionRetired(c.nextEventParams(modifiedBy), fieldID.Value(), optionID.Value())
	return c.applyAndRaise(event)
}

func (c *OnePagerConfiguration) nextEventParams(modifiedBy valueobjects.UserEmail) events.ModifyConfigurationParams {
	return events.ModifyConfigurationParams{
		ConfigID:   c.ID(),
		TenantID:   c.tenantID.Value(),
		Version:    c.Version() + 1,
		ModifiedBy: modifiedBy.Value(),
	}
}

func (c *OnePagerConfiguration) activeField(fieldID valueobjects.FieldID) (valueobjects.CustomField, error) {
	index := c.findFieldIndex(fieldID.Value())
	if index < 0 {
		return valueobjects.CustomField{}, ErrFieldNotFound
	}
	field := c.customFields[index]
	if !field.IsActive() {
		return valueobjects.CustomField{}, ErrFieldRetired
	}
	return field, nil
}

func (c *OnePagerConfiguration) findFieldIndex(fieldID string) int {
	for i, field := range c.customFields {
		if field.ID().Value() == fieldID {
			return i
		}
	}
	return -1
}

func (c *OnePagerConfiguration) isBuiltInIncluded(entryID string) bool {
	for _, ref := range c.displayOrder {
		if ref.Kind() == valueobjects.FieldRefKindBuiltIn && ref.RefID() == entryID {
			return true
		}
	}
	return false
}

func (c *OnePagerConfiguration) activeNameExists(name valueobjects.FieldName, excludeFieldID string) bool {
	return c.hasActiveCustomFieldNamed(name, excludeFieldID) || c.hasIncludedBuiltInLabeled(name)
}

func (c *OnePagerConfiguration) hasActiveCustomFieldNamed(name valueobjects.FieldName, excludeFieldID string) bool {
	for _, field := range c.customFields {
		if field.ID().Value() == excludeFieldID || !field.IsActive() {
			continue
		}
		if field.Name().EqualsIgnoreCase(name) {
			return true
		}
	}
	return false
}

func (c *OnePagerConfiguration) hasIncludedBuiltInLabeled(name valueobjects.FieldName) bool {
	for _, ref := range c.displayOrder {
		if ref.Kind() != valueobjects.FieldRefKindBuiltIn {
			continue
		}
		entry, found := catalog.LookupEntry(c.subjectType, ref.RefID())
		if !found {
			continue
		}
		if label, err := valueobjects.NewFieldName(entry.Label); err == nil && label.EqualsIgnoreCase(name) {
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

func optionsToEventData(options []valueobjects.SelectionOption) []events.SelectionOptionData {
	data := make([]events.SelectionOptionData, len(options))
	for i, option := range options {
		data[i] = events.SelectionOptionData{
			ID:     option.ID().Value(),
			Label:  option.Label().Value(),
			Active: option.IsActive(),
		}
	}
	return data
}

func refsToEventData(refs []valueobjects.FieldRef) []events.FieldRefData {
	data := make([]events.FieldRefData, len(refs))
	for i, ref := range refs {
		data[i] = events.FieldRefData{Kind: string(ref.Kind()), ID: ref.RefID()}
	}
	return data
}
