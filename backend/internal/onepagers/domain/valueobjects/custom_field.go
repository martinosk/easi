package valueobjects

import "errors"

var (
	ErrSelectionOptionRequired = errors.New("a selection field must define at least one option")
	ErrOptionsNotAllowed       = errors.New("only selection fields can define options")
	ErrDuplicateOptionLabel    = errors.New("option label already exists on this field")
	ErrNotSelectionField       = errors.New("field is not a selection field")
	ErrOptionNotFound          = errors.New("option not found on this field")
	ErrOptionAlreadyRetired    = errors.New("option is already retired")
	ErrLastActiveOption        = errors.New("cannot retire the last active option of a selection field")
)

type CustomField struct {
	id       FieldID
	name     FieldName
	dataType FieldType
	required bool
	helpText HelpText
	options  []SelectionOption
	active   bool
}

type CustomFieldParams struct {
	ID       FieldID
	Name     FieldName
	Type     FieldType
	Required bool
	HelpText HelpText
	Options  []SelectionOption
}

func NewCustomField(params CustomFieldParams) (CustomField, error) {
	if err := validateOptions(params.Type, params.Options); err != nil {
		return CustomField{}, err
	}
	return CustomField{
		id:       params.ID,
		name:     params.Name,
		dataType: params.Type,
		required: params.Required,
		helpText: params.HelpText,
		options:  copyOptions(params.Options),
		active:   true,
	}, nil
}

func validateOptions(fieldType FieldType, options []SelectionOption) error {
	if !fieldType.IsSelection() {
		return validateNonSelectionOptions(options)
	}
	return validateSelectionOptions(options)
}

func validateNonSelectionOptions(options []SelectionOption) error {
	if len(options) > 0 {
		return ErrOptionsNotAllowed
	}
	return nil
}

func validateSelectionOptions(options []SelectionOption) error {
	if len(options) == 0 {
		return ErrSelectionOptionRequired
	}
	if hasDuplicateOptionLabel(options) {
		return ErrDuplicateOptionLabel
	}
	return nil
}

func hasDuplicateOptionLabel(options []SelectionOption) bool {
	for i, option := range options {
		for _, previous := range options[:i] {
			if option.Label().EqualsIgnoreCase(previous.Label()) {
				return true
			}
		}
	}
	return false
}

func copyOptions(options []SelectionOption) []SelectionOption {
	copied := make([]SelectionOption, len(options))
	copy(copied, options)
	return copied
}

func (c CustomField) ID() FieldID {
	return c.id
}

func (c CustomField) Name() FieldName {
	return c.name
}

func (c CustomField) Type() FieldType {
	return c.dataType
}

func (c CustomField) IsRequired() bool {
	return c.required
}

func (c CustomField) HelpText() HelpText {
	return c.helpText
}

func (c CustomField) Options() []SelectionOption {
	return copyOptions(c.options)
}

func (c CustomField) IsActive() bool {
	return c.active
}

func (c CustomField) Renamed(name FieldName, helpText HelpText) CustomField {
	c.name = name
	c.helpText = helpText
	return c.withCopiedOptions()
}

func (c CustomField) WithRequirement(required bool) CustomField {
	c.required = required
	return c.withCopiedOptions()
}

func (c CustomField) Retired() CustomField {
	c.active = false
	return c.withCopiedOptions()
}

func (c CustomField) Reactivated() CustomField {
	c.active = true
	return c.withCopiedOptions()
}

func (c CustomField) WithAddedOption(option SelectionOption) (CustomField, error) {
	if !c.dataType.IsSelection() {
		return CustomField{}, ErrNotSelectionField
	}
	for _, existing := range c.options {
		if existing.IsActive() && existing.Label().EqualsIgnoreCase(option.Label()) {
			return CustomField{}, ErrDuplicateOptionLabel
		}
	}
	c.options = append(copyOptions(c.options), option)
	return c, nil
}

func (c CustomField) WithRetiredOption(optionID OptionID) (CustomField, error) {
	index := -1
	activeCount := 0
	for i, option := range c.options {
		if option.IsActive() {
			activeCount++
		}
		if option.ID() == optionID {
			index = i
		}
	}
	if index < 0 {
		return CustomField{}, ErrOptionNotFound
	}
	if !c.options[index].IsActive() {
		return CustomField{}, ErrOptionAlreadyRetired
	}
	if activeCount <= 1 {
		return CustomField{}, ErrLastActiveOption
	}
	options := copyOptions(c.options)
	options[index] = options[index].Retired()
	c.options = options
	return c, nil
}

func (c CustomField) HasActiveOption(optionID OptionID) bool {
	for _, option := range c.options {
		if option.ID() == optionID && option.IsActive() {
			return true
		}
	}
	return false
}

func (c CustomField) withCopiedOptions() CustomField {
	c.options = copyOptions(c.options)
	return c
}
