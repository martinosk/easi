package commands

type CreateOnePagerConfiguration struct {
	TenantID    string
	SubjectType string
	CreatedBy   string
}

func (c CreateOnePagerConfiguration) CommandName() string { return "CreateOnePagerConfiguration" }

type DefineCustomField struct {
	ConfigID     string
	Name         string
	FieldType    string
	Required     bool
	HelpText     string
	OptionLabels []string
	ModifiedBy   string
}

func (c DefineCustomField) CommandName() string     { return "DefineCustomField" }
func (c DefineCustomField) ConfigurationID() string { return c.ConfigID }
func (c DefineCustomField) ModifiedByEmail() string { return c.ModifiedBy }

type RenameCustomField struct {
	ConfigID      string
	FieldID       string
	Name          string
	HelpText      string
	RequestedType string
	ModifiedBy    string
}

func (c RenameCustomField) CommandName() string     { return "RenameCustomField" }
func (c RenameCustomField) ConfigurationID() string { return c.ConfigID }
func (c RenameCustomField) ModifiedByEmail() string { return c.ModifiedBy }

type ChangeCustomFieldRequirement struct {
	ConfigID   string
	FieldID    string
	Required   bool
	ModifiedBy string
}

func (c ChangeCustomFieldRequirement) CommandName() string     { return "ChangeCustomFieldRequirement" }
func (c ChangeCustomFieldRequirement) ConfigurationID() string { return c.ConfigID }
func (c ChangeCustomFieldRequirement) ModifiedByEmail() string { return c.ModifiedBy }

type RetireCustomField struct {
	ConfigID   string
	FieldID    string
	ModifiedBy string
}

func (c RetireCustomField) CommandName() string     { return "RetireCustomField" }
func (c RetireCustomField) ConfigurationID() string { return c.ConfigID }
func (c RetireCustomField) ModifiedByEmail() string { return c.ModifiedBy }

type ReactivateCustomField struct {
	ConfigID   string
	FieldID    string
	ModifiedBy string
}

func (c ReactivateCustomField) CommandName() string     { return "ReactivateCustomField" }
func (c ReactivateCustomField) ConfigurationID() string { return c.ConfigID }
func (c ReactivateCustomField) ModifiedByEmail() string { return c.ModifiedBy }

type IncludeBuiltInField struct {
	ConfigID   string
	EntryID    string
	ModifiedBy string
}

func (c IncludeBuiltInField) CommandName() string     { return "IncludeBuiltInField" }
func (c IncludeBuiltInField) ConfigurationID() string { return c.ConfigID }
func (c IncludeBuiltInField) ModifiedByEmail() string { return c.ModifiedBy }

type ExcludeBuiltInField struct {
	ConfigID   string
	EntryID    string
	ModifiedBy string
}

func (c ExcludeBuiltInField) CommandName() string     { return "ExcludeBuiltInField" }
func (c ExcludeBuiltInField) ConfigurationID() string { return c.ConfigID }
func (c ExcludeBuiltInField) ModifiedByEmail() string { return c.ModifiedBy }

type ChangeBuiltInFieldRequirement struct {
	ConfigID   string
	EntryID    string
	Required   bool
	ModifiedBy string
}

func (c ChangeBuiltInFieldRequirement) CommandName() string     { return "ChangeBuiltInFieldRequirement" }
func (c ChangeBuiltInFieldRequirement) ConfigurationID() string { return c.ConfigID }
func (c ChangeBuiltInFieldRequirement) ModifiedByEmail() string { return c.ModifiedBy }

type FieldRefInput struct {
	Kind string
	ID   string
}

type ReorderOnePagerFields struct {
	ConfigID   string
	Order      []FieldRefInput
	ModifiedBy string
}

func (c ReorderOnePagerFields) CommandName() string     { return "ReorderOnePagerFields" }
func (c ReorderOnePagerFields) ConfigurationID() string { return c.ConfigID }
func (c ReorderOnePagerFields) ModifiedByEmail() string { return c.ModifiedBy }

type AddSelectionOption struct {
	ConfigID   string
	FieldID    string
	Label      string
	ModifiedBy string
}

func (c AddSelectionOption) CommandName() string     { return "AddSelectionOption" }
func (c AddSelectionOption) ConfigurationID() string { return c.ConfigID }
func (c AddSelectionOption) ModifiedByEmail() string { return c.ModifiedBy }

type SetNumberFieldBounds struct {
	ConfigID   string
	FieldID    string
	Min        *float64
	Max        *float64
	ModifiedBy string
}

func (c SetNumberFieldBounds) CommandName() string     { return "SetNumberFieldBounds" }
func (c SetNumberFieldBounds) ConfigurationID() string { return c.ConfigID }
func (c SetNumberFieldBounds) ModifiedByEmail() string { return c.ModifiedBy }

type RetireSelectionOption struct {
	ConfigID   string
	FieldID    string
	OptionID   string
	ModifiedBy string
}

func (c RetireSelectionOption) CommandName() string     { return "RetireSelectionOption" }
func (c RetireSelectionOption) ConfigurationID() string { return c.ConfigID }
func (c RetireSelectionOption) ModifiedByEmail() string { return c.ModifiedBy }
