package valueobjects

import (
	"encoding/json"
	"errors"
	"fmt"
)

const ValueEnvelopeVersion = 1

var (
	ErrUnknownValueType        = errors.New("unknown field value type")
	ErrUnsupportedValueVersion = errors.New("unsupported field value version")
)

type ValueEnvelope struct {
	Type    string          `json:"type"`
	Version int             `json:"version"`
	Value   json.RawMessage `json:"value"`
}

type linkPayload struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type selectionPayload struct {
	OptionID string `json:"optionId"`
}

type contactPersonPayload struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company,omitempty"`
}

func NewValueEnvelope(value FieldValue) (ValueEnvelope, error) {
	payload, err := json.Marshal(envelopePayload(value))
	if err != nil {
		return ValueEnvelope{}, fmt.Errorf("marshal %s field value: %w", value.FieldTypeValue(), err)
	}
	return ValueEnvelope{Type: value.FieldTypeValue(), Version: ValueEnvelopeVersion, Value: payload}, nil
}

func envelopePayload(value FieldValue) interface{} {
	switch v := value.(type) {
	case TextValue:
		return v.Value()
	case NumberValue:
		return v.Value()
	case DateValue:
		return v.Value()
	case LinkValue:
		return linkPayload{Label: v.Label(), URL: v.URL().Value()}
	case SelectionValue:
		return selectionPayload{OptionID: v.OptionID().Value()}
	case ContactPerson:
		return contactPersonPayload{Name: v.Name(), Email: v.Email().Value(), Company: v.Company()}
	default:
		return nil
	}
}

var envelopeDecoders = map[string]func(json.RawMessage) (FieldValue, error){
	"text":      decodePayload(func(raw string) (FieldValue, error) { return NewTextValue(raw) }),
	"number":    decodePayload(func(raw float64) (FieldValue, error) { return NewNumberValue(raw) }),
	"date":      decodePayload(func(raw string) (FieldValue, error) { return NewDateValue(raw) }),
	"link":      decodePayload(func(p linkPayload) (FieldValue, error) { return NewLinkValue(p.Label, p.URL) }),
	"selection": decodePayload(func(p selectionPayload) (FieldValue, error) { return NewSelectionValue(p.OptionID) }),
	"contact-person": decodePayload(func(p contactPersonPayload) (FieldValue, error) {
		return NewContactPerson(ContactPersonParams(p))
	}),
}

func decodePayload[P any](construct func(P) (FieldValue, error)) func(json.RawMessage) (FieldValue, error) {
	return func(raw json.RawMessage) (FieldValue, error) {
		var payload P
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("unmarshal field value payload: %w", err)
		}
		return construct(payload)
	}
}

func FieldValueFromEnvelope(envelope ValueEnvelope) (FieldValue, error) {
	if envelope.Version != ValueEnvelopeVersion {
		return nil, ErrUnsupportedValueVersion
	}
	decode, found := envelopeDecoders[envelope.Type]
	if !found {
		return nil, ErrUnknownValueType
	}
	return decode(envelope.Value)
}
