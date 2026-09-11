package valueobjects

import "errors"

var (
	ErrSelectionOptionRequired = errors.New("a selection attribute must define at least one option")
	ErrOptionsNotAllowed       = errors.New("only selection attributes can define options")
	ErrDuplicateOptionLabel    = errors.New("option label already exists on this attribute")
	ErrNotSelectionAttribute   = errors.New("attribute is not a selection attribute")
	ErrOptionNotFound          = errors.New("option not found on this attribute")
	ErrOptionAlreadyRetired    = errors.New("option is already retired")
	ErrLastActiveOption        = errors.New("cannot retire the last active option of a selection attribute")
	ErrBoundsNotAllowed        = errors.New("only number attributes can define bounds")
	ErrMinExceedsMax           = errors.New("minimum bound must not exceed maximum bound")
)

type SubjectAttribute struct {
	id       AttributeID
	name     AttributeName
	dataType AttributeType
	helpText HelpText
	options  []AttributeOption
	active   bool
	min      *float64
	max      *float64
}

type SubjectAttributeParams struct {
	ID       AttributeID
	Name     AttributeName
	Type     AttributeType
	HelpText HelpText
	Options  []AttributeOption
	Min      *float64
	Max      *float64
}

func NewSubjectAttribute(params SubjectAttributeParams) (SubjectAttribute, error) {
	if err := validateOptions(params.Type, params.Options); err != nil {
		return SubjectAttribute{}, err
	}
	if err := validateBounds(params.Type, params.Min, params.Max); err != nil {
		return SubjectAttribute{}, err
	}
	return SubjectAttribute{
		id:       params.ID,
		name:     params.Name,
		dataType: params.Type,
		helpText: params.HelpText,
		options:  copyOptions(params.Options),
		active:   true,
		min:      copyFloatPtr(params.Min),
		max:      copyFloatPtr(params.Max),
	}, nil
}

func validateOptions(attributeType AttributeType, options []AttributeOption) error {
	if !attributeType.IsSelection() {
		if len(options) > 0 {
			return ErrOptionsNotAllowed
		}
		return nil
	}
	if len(options) == 0 {
		return ErrSelectionOptionRequired
	}
	if hasDuplicateOptionLabel(options) {
		return ErrDuplicateOptionLabel
	}
	return nil
}

func hasDuplicateOptionLabel(options []AttributeOption) bool {
	for i, option := range options {
		for _, previous := range options[:i] {
			if option.Label().EqualsIgnoreCase(previous.Label()) {
				return true
			}
		}
	}
	return false
}

func validateBounds(attributeType AttributeType, min, max *float64) error {
	if !attributeType.IsNumber() {
		if hasAnyBound(min, max) {
			return ErrBoundsNotAllowed
		}
		return nil
	}
	if minExceedsMax(min, max) {
		return ErrMinExceedsMax
	}
	return nil
}

func hasAnyBound(min, max *float64) bool {
	return min != nil || max != nil
}

func minExceedsMax(min, max *float64) bool {
	return min != nil && max != nil && *min > *max
}

func copyFloatPtr(v *float64) *float64 {
	if v == nil {
		return nil
	}
	copied := *v
	return &copied
}

func copyOptions(options []AttributeOption) []AttributeOption {
	copied := make([]AttributeOption, len(options))
	copy(copied, options)
	return copied
}

func (a SubjectAttribute) ID() AttributeID {
	return a.id
}

func (a SubjectAttribute) Name() AttributeName {
	return a.name
}

func (a SubjectAttribute) Type() AttributeType {
	return a.dataType
}

func (a SubjectAttribute) HelpText() HelpText {
	return a.helpText
}

func (a SubjectAttribute) Options() []AttributeOption {
	return copyOptions(a.options)
}

func (a SubjectAttribute) IsActive() bool {
	return a.active
}

func (a SubjectAttribute) Min() *float64 {
	return copyFloatPtr(a.min)
}

func (a SubjectAttribute) Max() *float64 {
	return copyFloatPtr(a.max)
}

func (a SubjectAttribute) Renamed(name AttributeName, helpText HelpText) SubjectAttribute {
	a.name = name
	a.helpText = helpText
	return a.withCopiedOptions()
}

func (a SubjectAttribute) WithBounds(min, max *float64) (SubjectAttribute, error) {
	if err := validateBounds(a.dataType, min, max); err != nil {
		return SubjectAttribute{}, err
	}
	a.min = copyFloatPtr(min)
	a.max = copyFloatPtr(max)
	return a.withCopiedOptions(), nil
}

func (a SubjectAttribute) Retired() SubjectAttribute {
	a.active = false
	return a.withCopiedOptions()
}

func (a SubjectAttribute) Reactivated() SubjectAttribute {
	a.active = true
	return a.withCopiedOptions()
}

func (a SubjectAttribute) WithAddedOption(option AttributeOption) (SubjectAttribute, error) {
	if !a.dataType.IsSelection() {
		return SubjectAttribute{}, ErrNotSelectionAttribute
	}
	for _, existing := range a.options {
		if existing.IsActive() && existing.Label().EqualsIgnoreCase(option.Label()) {
			return SubjectAttribute{}, ErrDuplicateOptionLabel
		}
	}
	a.options = append(copyOptions(a.options), option)
	return a, nil
}

func (a SubjectAttribute) WithRetiredOption(optionID OptionID) (SubjectAttribute, error) {
	index := -1
	activeCount := 0
	for i, option := range a.options {
		if option.IsActive() {
			activeCount++
		}
		if option.ID() == optionID {
			index = i
		}
	}
	if index < 0 {
		return SubjectAttribute{}, ErrOptionNotFound
	}
	if !a.options[index].IsActive() {
		return SubjectAttribute{}, ErrOptionAlreadyRetired
	}
	if activeCount <= 1 {
		return SubjectAttribute{}, ErrLastActiveOption
	}
	options := copyOptions(a.options)
	options[index] = options[index].Retired()
	a.options = options
	return a, nil
}

func (a SubjectAttribute) HasActiveOption(optionID OptionID) bool {
	for _, option := range a.options {
		if option.ID() == optionID && option.IsActive() {
			return true
		}
	}
	return false
}

func (a SubjectAttribute) withCopiedOptions() SubjectAttribute {
	a.options = copyOptions(a.options)
	return a
}
