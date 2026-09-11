package readmodels

const (
	builtInRefKind = "builtIn"
	customRefKind  = "custom"
)

type FieldRefRecord struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type FieldRequirementRecord struct {
	ID       string `json:"id"`
	Required bool   `json:"required"`
}

type ConfigurationDocument struct {
	CustomFields  []FieldRequirementRecord `json:"customFields"`
	BuiltInFields []FieldRequirementRecord `json:"builtInFields,omitempty"`
	DisplayOrder  []FieldRefRecord         `json:"displayOrder"`
}

func (d ConfigurationDocument) BuiltInRequired(entryID string) bool {
	return requiredIn(d.BuiltInFields, entryID)
}

func (d ConfigurationDocument) CustomFieldRequired(fieldID string) bool {
	return requiredIn(d.CustomFields, fieldID)
}

func requiredIn(records []FieldRequirementRecord, id string) bool {
	for _, record := range records {
		if record.ID == id {
			return record.Required
		}
	}
	return false
}

func (d ConfigurationDocument) IncludedCustomFieldIDs() []string {
	return d.includedIDs(customRefKind, func(string) bool { return true })
}

func (d ConfigurationDocument) RequiredCustomFieldIDs() []string {
	return d.includedIDs(customRefKind, d.CustomFieldRequired)
}

func (d ConfigurationDocument) IncludedBuiltInEntryIDs() []string {
	return d.includedIDs(builtInRefKind, func(string) bool { return true })
}

func (d ConfigurationDocument) RequiredBuiltInEntryIDs() []string {
	return d.includedIDs(builtInRefKind, d.BuiltInRequired)
}

func (d ConfigurationDocument) includedIDs(kind string, accept func(string) bool) []string {
	ids := make([]string, 0, len(d.DisplayOrder))
	for _, ref := range d.DisplayOrder {
		if ref.Kind == kind && accept(ref.ID) {
			ids = append(ids, ref.ID)
		}
	}
	return ids
}
