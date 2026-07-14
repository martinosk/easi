package events

const (
	TypeOnePagerConfigurationCreated   = "OnePagerConfigurationCreated"
	TypeCustomFieldDefined             = "CustomFieldDefined"
	TypeCustomFieldRenamed             = "CustomFieldRenamed"
	TypeCustomFieldRequirementChanged  = "CustomFieldRequirementChanged"
	TypeCustomFieldRetired             = "CustomFieldRetired"
	TypeCustomFieldReactivated         = "CustomFieldReactivated"
	TypeBuiltInFieldIncluded           = "BuiltInFieldIncluded"
	TypeBuiltInFieldExcluded           = "BuiltInFieldExcluded"
	TypeBuiltInFieldRequirementChanged = "BuiltInFieldRequirementChanged"
	TypeOnePagerFieldsReordered        = "OnePagerFieldsReordered"
	TypeSelectionOptionAdded           = "SelectionOptionAdded"
	TypeSelectionOptionRetired         = "SelectionOptionRetired"
	TypeNumberFieldBoundsChanged       = "NumberFieldBoundsChanged"
	TypeFieldValueRecorded             = "FieldValueRecorded"
	TypeFieldValueCleared              = "FieldValueCleared"
	TypeOnePagerFactsArchived          = "OnePagerFactsArchived"
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
		TypeBuiltInFieldRequirementChanged,
		TypeOnePagerFieldsReordered,
		TypeSelectionOptionAdded,
		TypeSelectionOptionRetired,
		TypeNumberFieldBoundsChanged,
	}
}

func FactsEventTypes() []string {
	return []string{
		TypeFieldValueRecorded,
		TypeFieldValueCleared,
		TypeOnePagerFactsArchived,
	}
}
