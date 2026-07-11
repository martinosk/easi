package readmodels

import "easi/backend/internal/onepagers/domain/valueobjects"

func (f CustomFieldRecord) RetiredOptionReferenced(value *valueobjects.ValueEnvelope) bool {
	option, found := f.selectedOption(value)
	return found && !option.Active
}

func (f CustomFieldRecord) SelectionOptionLabel(value *valueobjects.ValueEnvelope) (string, bool) {
	option, found := f.selectedOption(value)
	if !found {
		return "", false
	}
	return option.Label, true
}

func (f CustomFieldRecord) selectedOption(value *valueobjects.ValueEnvelope) (OptionRecord, bool) {
	if value == nil {
		return OptionRecord{}, false
	}
	decoded, err := valueobjects.FieldValueFromEnvelope(*value)
	if err != nil {
		return OptionRecord{}, false
	}
	selection, ok := decoded.(valueobjects.SelectionValue)
	if !ok {
		return OptionRecord{}, false
	}
	optionID := selection.OptionID().Value()
	for _, option := range f.Options {
		if option.ID == optionID {
			return option, true
		}
	}
	return OptionRecord{}, false
}

func (d ConfigurationDocument) CustomField(fieldID string) (CustomFieldRecord, bool) {
	for _, field := range d.CustomFields {
		if field.ID == fieldID {
			return field, true
		}
	}
	return CustomFieldRecord{}, false
}
