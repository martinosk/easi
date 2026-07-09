package publishedlanguage

const (
	OnePagerConfigurationCreated  = "OnePagerConfigurationCreated"
	CustomFieldDefined            = "CustomFieldDefined"
	CustomFieldRenamed            = "CustomFieldRenamed"
	CustomFieldRequirementChanged = "CustomFieldRequirementChanged"
	CustomFieldRetired            = "CustomFieldRetired"
	CustomFieldReactivated        = "CustomFieldReactivated"
	BuiltInFieldIncluded          = "BuiltInFieldIncluded"
	BuiltInFieldExcluded          = "BuiltInFieldExcluded"
	OnePagerFieldsReordered       = "OnePagerFieldsReordered"
	SelectionOptionAdded          = "SelectionOptionAdded"
	SelectionOptionRetired        = "SelectionOptionRetired"
)

func AllEventTypes() []string {
	return []string{
		OnePagerConfigurationCreated,
		CustomFieldDefined,
		CustomFieldRenamed,
		CustomFieldRequirementChanged,
		CustomFieldRetired,
		CustomFieldReactivated,
		BuiltInFieldIncluded,
		BuiltInFieldExcluded,
		OnePagerFieldsReordered,
		SelectionOptionAdded,
		SelectionOptionRetired,
	}
}
