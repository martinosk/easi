package readmodels

const builtInRefKind = "builtIn"

type OptionRecord struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

type CustomFieldRecord struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Required bool           `json:"required"`
	HelpText string         `json:"helpText"`
	Active   bool           `json:"active"`
	Options  []OptionRecord `json:"options,omitempty"`
	Min      *float64       `json:"min,omitempty"`
	Max      *float64       `json:"max,omitempty"`
}

type FieldRefRecord struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type BuiltInFieldRecord struct {
	ID       string `json:"id"`
	Required bool   `json:"required"`
}

type ConfigurationDocument struct {
	CustomFields  []CustomFieldRecord  `json:"customFields"`
	BuiltInFields []BuiltInFieldRecord `json:"builtInFields,omitempty"`
	DisplayOrder  []FieldRefRecord     `json:"displayOrder"`
}

func (d ConfigurationDocument) BuiltInRequired(entryID string) bool {
	for _, builtIn := range d.BuiltInFields {
		if builtIn.ID == entryID {
			return builtIn.Required
		}
	}
	return false
}

func (d ConfigurationDocument) ActiveRequiredCustomFieldIDs() []string {
	var fieldIDs []string
	for _, field := range d.CustomFields {
		if field.Active && field.Required {
			fieldIDs = append(fieldIDs, field.ID)
		}
	}
	return fieldIDs
}

func (d ConfigurationDocument) IncludedBuiltInEntryIDs() []string {
	entryIDs := make([]string, 0, len(d.DisplayOrder))
	for _, ref := range d.DisplayOrder {
		if ref.Kind == builtInRefKind {
			entryIDs = append(entryIDs, ref.ID)
		}
	}
	return entryIDs
}

func (d ConfigurationDocument) RequiredBuiltInEntryIDs() []string {
	var entryIDs []string
	for _, ref := range d.DisplayOrder {
		if ref.Kind == builtInRefKind && d.BuiltInRequired(ref.ID) {
			entryIDs = append(entryIDs, ref.ID)
		}
	}
	return entryIDs
}
