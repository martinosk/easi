package events

type OnePagerFieldsReordered struct {
	ConfigurationEventBase
	Order []FieldRefData `json:"order"`
}

func NewOnePagerFieldsReordered(params ModifyConfigurationParams, order []FieldRefData) OnePagerFieldsReordered {
	return OnePagerFieldsReordered{
		ConfigurationEventBase: newConfigurationEventBase(params),
		Order:                  order,
	}
}

func (e OnePagerFieldsReordered) EventType() string {
	return "OnePagerFieldsReordered"
}

func (e OnePagerFieldsReordered) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["order"] = e.Order
	return data
}

type SelectionOptionAdded struct {
	ConfigurationEventBase
	FieldID  string `json:"fieldId"`
	OptionID string `json:"optionId"`
	Label    string `json:"label"`
}

func NewSelectionOptionAdded(params ModifyConfigurationParams, fieldID, optionID, label string) SelectionOptionAdded {
	return SelectionOptionAdded{
		ConfigurationEventBase: newConfigurationEventBase(params),
		FieldID:                fieldID,
		OptionID:               optionID,
		Label:                  label,
	}
}

func (e SelectionOptionAdded) EventType() string {
	return "SelectionOptionAdded"
}

func (e SelectionOptionAdded) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["fieldId"] = e.FieldID
	data["optionId"] = e.OptionID
	data["label"] = e.Label
	return data
}

type SelectionOptionRetired struct {
	ConfigurationEventBase
	FieldID  string `json:"fieldId"`
	OptionID string `json:"optionId"`
}

func NewSelectionOptionRetired(params ModifyConfigurationParams, fieldID, optionID string) SelectionOptionRetired {
	return SelectionOptionRetired{
		ConfigurationEventBase: newConfigurationEventBase(params),
		FieldID:                fieldID,
		OptionID:               optionID,
	}
}

func (e SelectionOptionRetired) EventType() string {
	return "SelectionOptionRetired"
}

func (e SelectionOptionRetired) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["fieldId"] = e.FieldID
	data["optionId"] = e.OptionID
	return data
}
