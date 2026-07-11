package events

const (
	TypeOnePagerConfigurationCreated  = "OnePagerConfigurationCreated"
	TypeCustomFieldDefined            = "CustomFieldDefined"
	TypeCustomFieldRenamed            = "CustomFieldRenamed"
	TypeCustomFieldRequirementChanged = "CustomFieldRequirementChanged"
	TypeCustomFieldRetired            = "CustomFieldRetired"
	TypeCustomFieldReactivated        = "CustomFieldReactivated"
	TypeBuiltInFieldIncluded          = "BuiltInFieldIncluded"
	TypeBuiltInFieldExcluded          = "BuiltInFieldExcluded"
	TypeOnePagerFieldsReordered       = "OnePagerFieldsReordered"
	TypeSelectionOptionAdded          = "SelectionOptionAdded"
	TypeSelectionOptionRetired        = "SelectionOptionRetired"
	TypeFieldValueRecorded            = "FieldValueRecorded"
	TypeFieldValueCleared             = "FieldValueCleared"
	TypeOnePagerFactsArchived         = "OnePagerFactsArchived"
)

func ConfigurationEventTypes() []string {
	return []string{
		TypeOnePagerConfigurationCreated,
		TypeCustomFieldDefined,
		TypeCustomFieldRenamed,
		TypeCustomFieldRequirementChanged,
		TypeCustomFieldRetired,
		TypeCustomFieldReactivated,
		TypeBuiltInFieldIncluded,
		TypeBuiltInFieldExcluded,
		TypeOnePagerFieldsReordered,
		TypeSelectionOptionAdded,
		TypeSelectionOptionRetired,
	}
}

func FactsEventTypes() []string {
	return []string{
		TypeFieldValueRecorded,
		TypeFieldValueCleared,
		TypeOnePagerFactsArchived,
	}
}
