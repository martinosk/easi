package events

type CustomFieldDefined struct {
	ConfigurationEventBase
	FieldID   string                `json:"fieldId"`
	Name      string                `json:"name"`
	FieldType string                `json:"fieldType"`
	Required  bool                  `json:"required"`
	HelpText  string                `json:"helpText"`
	Options   []SelectionOptionData `json:"options"`
}

type CustomFieldData struct {
	FieldID   string
	Name      string
	FieldType string
	Required  bool
	HelpText  string
	Options   []SelectionOptionData
}

func NewCustomFieldDefined(params ModifyConfigurationParams, field CustomFieldData) CustomFieldDefined {
	return CustomFieldDefined{
		ConfigurationEventBase: newConfigurationEventBase(params),
		FieldID:                field.FieldID,
		Name:                   field.Name,
		FieldType:              field.FieldType,
		Required:               field.Required,
		HelpText:               field.HelpText,
		Options:                field.Options,
	}
}

func (e CustomFieldDefined) EventType() string {
	return "CustomFieldDefined"
}

func (e CustomFieldDefined) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["fieldId"] = e.FieldID
	data["name"] = e.Name
	data["fieldType"] = e.FieldType
	data["required"] = e.Required
	data["helpText"] = e.HelpText
	data["options"] = e.Options
	return data
}

type CustomFieldRenamed struct {
	ConfigurationEventBase
	FieldID     string `json:"fieldId"`
	NewName     string `json:"newName"`
	NewHelpText string `json:"newHelpText"`
}

type FieldRenameData struct {
	FieldID     string
	NewName     string
	NewHelpText string
}

func NewCustomFieldRenamed(params ModifyConfigurationParams, rename FieldRenameData) CustomFieldRenamed {
	return CustomFieldRenamed{
		ConfigurationEventBase: newConfigurationEventBase(params),
		FieldID:                rename.FieldID,
		NewName:                rename.NewName,
		NewHelpText:            rename.NewHelpText,
	}
}

func (e CustomFieldRenamed) EventType() string {
	return "CustomFieldRenamed"
}

func (e CustomFieldRenamed) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["fieldId"] = e.FieldID
	data["newName"] = e.NewName
	data["newHelpText"] = e.NewHelpText
	return data
}

type CustomFieldRequirementChanged struct {
	ConfigurationEventBase
	FieldID  string `json:"fieldId"`
	Required bool   `json:"required"`
}

func NewCustomFieldRequirementChanged(params ModifyConfigurationParams, fieldID string, required bool) CustomFieldRequirementChanged {
	return CustomFieldRequirementChanged{
		ConfigurationEventBase: newConfigurationEventBase(params),
		FieldID:                fieldID,
		Required:               required,
	}
}

func (e CustomFieldRequirementChanged) EventType() string {
	return "CustomFieldRequirementChanged"
}

func (e CustomFieldRequirementChanged) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["fieldId"] = e.FieldID
	data["required"] = e.Required
	return data
}

type CustomFieldRetired struct {
	ConfigurationEventBase
	FieldID string `json:"fieldId"`
}

func NewCustomFieldRetired(params ModifyConfigurationParams, fieldID string) CustomFieldRetired {
	return CustomFieldRetired{
		ConfigurationEventBase: newConfigurationEventBase(params),
		FieldID:                fieldID,
	}
}

func (e CustomFieldRetired) EventType() string {
	return "CustomFieldRetired"
}

func (e CustomFieldRetired) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["fieldId"] = e.FieldID
	return data
}

type CustomFieldReactivated struct {
	ConfigurationEventBase
	FieldID string `json:"fieldId"`
}

func NewCustomFieldReactivated(params ModifyConfigurationParams, fieldID string) CustomFieldReactivated {
	return CustomFieldReactivated{
		ConfigurationEventBase: newConfigurationEventBase(params),
		FieldID:                fieldID,
	}
}

func (e CustomFieldReactivated) EventType() string {
	return "CustomFieldReactivated"
}

func (e CustomFieldReactivated) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["fieldId"] = e.FieldID
	return data
}

type NumberFieldBoundsChanged struct {
	ConfigurationEventBase
	FieldID string   `json:"fieldId"`
	Min     *float64 `json:"min,omitempty"`
	Max     *float64 `json:"max,omitempty"`
}

func NewNumberFieldBoundsChanged(params ModifyConfigurationParams, fieldID string, min, max *float64) NumberFieldBoundsChanged {
	return NumberFieldBoundsChanged{
		ConfigurationEventBase: newConfigurationEventBase(params),
		FieldID:                fieldID,
		Min:                    min,
		Max:                    max,
	}
}

func (e NumberFieldBoundsChanged) EventType() string {
	return "NumberFieldBoundsChanged"
}

func (e NumberFieldBoundsChanged) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["fieldId"] = e.FieldID
	data["min"] = e.Min
	data["max"] = e.Max
	return data
}
