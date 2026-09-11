package commands

type CreateOnePagerConfiguration struct {
	TenantID    string
	SubjectType string
	CreatedBy   string
}

func (c CreateOnePagerConfiguration) CommandName() string { return "CreateOnePagerConfiguration" }

type IncludeCustomField struct {
	ConfigID   string
	FieldID    string
	ModifiedBy string
}

func (c IncludeCustomField) CommandName() string     { return "IncludeCustomField" }
func (c IncludeCustomField) ConfigurationID() string { return c.ConfigID }
func (c IncludeCustomField) ModifiedByEmail() string { return c.ModifiedBy }

type ExcludeCustomField struct {
	ConfigID   string
	FieldID    string
	ModifiedBy string
}

func (c ExcludeCustomField) CommandName() string     { return "ExcludeCustomField" }
func (c ExcludeCustomField) ConfigurationID() string { return c.ConfigID }
func (c ExcludeCustomField) ModifiedByEmail() string { return c.ModifiedBy }

type ChangeCustomFieldRequirement struct {
	ConfigID   string
	FieldID    string
	Required   bool
	ModifiedBy string
}

func (c ChangeCustomFieldRequirement) CommandName() string     { return "ChangeCustomFieldRequirement" }
func (c ChangeCustomFieldRequirement) ConfigurationID() string { return c.ConfigID }
func (c ChangeCustomFieldRequirement) ModifiedByEmail() string { return c.ModifiedBy }

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
