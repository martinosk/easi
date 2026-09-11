package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrAttributeNameEmpty   = errors.New("attribute name cannot be empty")
	ErrAttributeNameTooLong = errors.New("attribute name cannot exceed 100 characters")
	ErrHelpTextTooLong      = errors.New("help text cannot exceed 500 characters")
	ErrOptionLabelEmpty     = errors.New("option label cannot be empty")
	ErrOptionLabelTooLong   = errors.New("option label cannot exceed 100 characters")
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

type AttributeName struct {
	value string
}

func NewAttributeName(value string) (AttributeName, error) {
	trimmed, err := validatedLabel(value, ErrAttributeNameEmpty, ErrAttributeNameTooLong)
	if err != nil {
		return AttributeName{}, err
	}
	return AttributeName{value: trimmed}, nil
}

func (a AttributeName) Value() string {
	return a.value
}

func (a AttributeName) EqualsIgnoreCase(other AttributeName) bool {
	return strings.EqualFold(a.value, other.value)
}

func (a AttributeName) Equals(other domain.ValueObject) bool {
	if o, ok := other.(AttributeName); ok {
		return a.value == o.value
	}
	return false
}

func (a AttributeName) String() string {
	return a.value
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
