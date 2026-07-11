package valueobjects

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

const (
	MaxTextValueLength      = 2000
	MaxLinkLabelLength      = 200
	MaxContactNameLength    = 200
	MaxContactCompanyLength = 200

	ISODateLayout = "2006-01-02"
)

var (
	ErrTextValueEmpty        = errors.New("text value cannot be empty")
	ErrTextValueTooLong      = errors.New("text value exceeds maximum length of 2000 characters")
	ErrNumberValueNotFinite  = errors.New("number value must be a finite number")
	ErrDateValueInvalid      = errors.New("date value must be an ISO date (YYYY-MM-DD)")
	ErrLinkLabelEmpty        = errors.New("link label cannot be empty")
	ErrLinkLabelTooLong      = errors.New("link label exceeds maximum length of 200 characters")
	ErrContactNameEmpty      = errors.New("contact person name cannot be empty")
	ErrContactNameTooLong    = errors.New("contact person name exceeds maximum length of 200 characters")
	ErrContactCompanyTooLong = errors.New("contact person company exceeds maximum length of 200 characters")
)

type FieldValue interface {
	domain.ValueObject
	FieldTypeValue() string
}

func DisplayText(value FieldValue) string {
	switch v := value.(type) {
	case TextValue:
		return v.Value()
	case NumberValue:
		return strconv.FormatFloat(v.Value(), 'f', -1, 64)
	case DateValue:
		return v.Value()
	case LinkValue:
		return v.Label()
	case SelectionValue:
		return v.OptionID().Value()
	case ContactPerson:
		return v.Name()
	default:
		return ""
	}
}

type TextValue struct {
	value string
}

func NewTextValue(value string) (TextValue, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return TextValue{}, ErrTextValueEmpty
	}
	if len(trimmed) > MaxTextValueLength {
		return TextValue{}, ErrTextValueTooLong
	}
	return TextValue{value: trimmed}, nil
}

func (v TextValue) Value() string {
	return v.value
}

func (v TextValue) FieldTypeValue() string {
	return "text"
}

func (v TextValue) Equals(other domain.ValueObject) bool {
	if o, ok := other.(TextValue); ok {
		return v.value == o.value
	}
	return false
}

type NumberValue struct {
	value float64
}

func NewNumberValue(value float64) (NumberValue, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return NumberValue{}, ErrNumberValueNotFinite
	}
	return NumberValue{value: value}, nil
}

func (v NumberValue) Value() float64 {
	return v.value
}

func (v NumberValue) FieldTypeValue() string {
	return "number"
}

func (v NumberValue) Equals(other domain.ValueObject) bool {
	if o, ok := other.(NumberValue); ok {
		return v.value == o.value
	}
	return false
}

type DateValue struct {
	value string
}

func NewDateValue(value string) (DateValue, error) {
	parsed, err := time.Parse(ISODateLayout, value)
	if err != nil {
		return DateValue{}, ErrDateValueInvalid
	}
	return DateValue{value: parsed.Format(ISODateLayout)}, nil
}

func (v DateValue) Value() string {
	return v.value
}

func (v DateValue) FieldTypeValue() string {
	return "date"
}

func (v DateValue) Equals(other domain.ValueObject) bool {
	if o, ok := other.(DateValue); ok {
		return v.value == o.value
	}
	return false
}

type LinkValue struct {
	label string
	url   sharedvo.URL
}

func NewLinkValue(label, rawURL string) (LinkValue, error) {
	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		return LinkValue{}, ErrLinkLabelEmpty
	}
	if len(trimmed) > MaxLinkLabelLength {
		return LinkValue{}, ErrLinkLabelTooLong
	}
	url, err := sharedvo.NewURL(rawURL)
	if err != nil {
		return LinkValue{}, err
	}
	return LinkValue{label: trimmed, url: url}, nil
}

func (v LinkValue) Label() string {
	return v.label
}

func (v LinkValue) URL() sharedvo.URL {
	return v.url
}

func (v LinkValue) FieldTypeValue() string {
	return "link"
}

func (v LinkValue) Equals(other domain.ValueObject) bool {
	if o, ok := other.(LinkValue); ok {
		return v.label == o.label && v.url.Equals(o.url)
	}
	return false
}

type SelectionValue struct {
	optionID OptionID
}

func NewSelectionValue(optionID string) (SelectionValue, error) {
	id, err := NewOptionIDFromString(optionID)
	if err != nil {
		return SelectionValue{}, err
	}
	return SelectionValue{optionID: id}, nil
}

func (v SelectionValue) OptionID() OptionID {
	return v.optionID
}

func (v SelectionValue) FieldTypeValue() string {
	return "selection"
}

func (v SelectionValue) Equals(other domain.ValueObject) bool {
	if o, ok := other.(SelectionValue); ok {
		return v.optionID == o.optionID
	}
	return false
}

type ContactPerson struct {
	name    string
	email   UserEmail
	company string
}

type ContactPersonParams struct {
	Name    string
	Email   string
	Company string
}

func NewContactPerson(params ContactPersonParams) (ContactPerson, error) {
	trimmedName := strings.TrimSpace(params.Name)
	if trimmedName == "" {
		return ContactPerson{}, ErrContactNameEmpty
	}
	if len(trimmedName) > MaxContactNameLength {
		return ContactPerson{}, ErrContactNameTooLong
	}
	contactEmail, err := NewUserEmail(params.Email)
	if err != nil {
		return ContactPerson{}, err
	}
	trimmedCompany := strings.TrimSpace(params.Company)
	if len(trimmedCompany) > MaxContactCompanyLength {
		return ContactPerson{}, ErrContactCompanyTooLong
	}
	return ContactPerson{name: trimmedName, email: contactEmail, company: trimmedCompany}, nil
}

func (v ContactPerson) Name() string {
	return v.name
}

func (v ContactPerson) Email() UserEmail {
	return v.email
}

func (v ContactPerson) Company() string {
	return v.company
}

func (v ContactPerson) FieldTypeValue() string {
	return "contact-person"
}

func (v ContactPerson) Equals(other domain.ValueObject) bool {
	if o, ok := other.(ContactPerson); ok {
		return v.name == o.name && v.email == o.email && v.company == o.company
	}
	return false
}
