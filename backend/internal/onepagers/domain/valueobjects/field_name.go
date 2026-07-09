package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrFieldNameEmpty     = errors.New("field name cannot be empty")
	ErrFieldNameTooLong   = errors.New("field name cannot exceed 100 characters")
	ErrHelpTextTooLong    = errors.New("help text cannot exceed 500 characters")
	ErrOptionLabelEmpty   = errors.New("option label cannot be empty")
	ErrOptionLabelTooLong = errors.New("option label cannot exceed 100 characters")
)

func validatedLabel(value string, emptyErr, tooLongErr error) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", emptyErr
	}
	if len(trimmed) > 100 {
		return "", tooLongErr
	}
	return trimmed, nil
}

type FieldName struct {
	value string
}

func NewFieldName(value string) (FieldName, error) {
	trimmed, err := validatedLabel(value, ErrFieldNameEmpty, ErrFieldNameTooLong)
	if err != nil {
		return FieldName{}, err
	}
	return FieldName{value: trimmed}, nil
}

func (f FieldName) Value() string {
	return f.value
}

func (f FieldName) EqualsIgnoreCase(other FieldName) bool {
	return strings.EqualFold(f.value, other.value)
}

func (f FieldName) Equals(other domain.ValueObject) bool {
	if o, ok := other.(FieldName); ok {
		return f.value == o.value
	}
	return false
}

func (f FieldName) String() string {
	return f.value
}

type HelpText struct {
	value string
}

func NewHelpText(value string) (HelpText, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > 500 {
		return HelpText{}, ErrHelpTextTooLong
	}
	return HelpText{value: trimmed}, nil
}

func (h HelpText) Value() string {
	return h.value
}

func (h HelpText) Equals(other domain.ValueObject) bool {
	if o, ok := other.(HelpText); ok {
		return h.value == o.value
	}
	return false
}

func (h HelpText) String() string {
	return h.value
}

type OptionLabel struct {
	value string
}

func NewOptionLabel(value string) (OptionLabel, error) {
	trimmed, err := validatedLabel(value, ErrOptionLabelEmpty, ErrOptionLabelTooLong)
	if err != nil {
		return OptionLabel{}, err
	}
	return OptionLabel{value: trimmed}, nil
}

func (o OptionLabel) Value() string {
	return o.value
}

func (o OptionLabel) EqualsIgnoreCase(other OptionLabel) bool {
	return strings.EqualFold(o.value, other.value)
}

func (o OptionLabel) Equals(other domain.ValueObject) bool {
	if other2, ok := other.(OptionLabel); ok {
		return o.value == other2.value
	}
	return false
}

func (o OptionLabel) String() string {
	return o.value
}
