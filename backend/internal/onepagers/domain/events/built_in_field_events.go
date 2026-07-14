package events

type BuiltInFieldIncluded struct {
	ConfigurationEventBase
	EntryID string `json:"entryId"`
}

func NewBuiltInFieldIncluded(params ModifyConfigurationParams, entryID string) BuiltInFieldIncluded {
	return BuiltInFieldIncluded{
		ConfigurationEventBase: newConfigurationEventBase(params),
		EntryID:                entryID,
	}
}

func (e BuiltInFieldIncluded) EventType() string {
	return "BuiltInFieldIncluded"
}

func (e BuiltInFieldIncluded) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["entryId"] = e.EntryID
	return data
}

type BuiltInFieldRequirementChanged struct {
	ConfigurationEventBase
	EntryID  string `json:"entryId"`
	Required bool   `json:"required"`
}

func NewBuiltInFieldRequirementChanged(params ModifyConfigurationParams, entryID string, required bool) BuiltInFieldRequirementChanged {
	return BuiltInFieldRequirementChanged{
		ConfigurationEventBase: newConfigurationEventBase(params),
		EntryID:                entryID,
		Required:               required,
	}
}

func (e BuiltInFieldRequirementChanged) EventType() string {
	return "BuiltInFieldRequirementChanged"
}

func (e BuiltInFieldRequirementChanged) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["entryId"] = e.EntryID
	data["required"] = e.Required
	return data
}

type BuiltInFieldExcluded struct {
	ConfigurationEventBase
	EntryID string `json:"entryId"`
}

func NewBuiltInFieldExcluded(params ModifyConfigurationParams, entryID string) BuiltInFieldExcluded {
	return BuiltInFieldExcluded{
		ConfigurationEventBase: newConfigurationEventBase(params),
		EntryID:                entryID,
	}
}

func (e BuiltInFieldExcluded) EventType() string {
	return "BuiltInFieldExcluded"
}

func (e BuiltInFieldExcluded) EventData() map[string]interface{} {
	data := e.baseEventData()
	data["entryId"] = e.EntryID
	return data
}
