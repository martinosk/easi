package valueobjects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustFieldValues(t *testing.T) []FieldValue {
	t.Helper()
	text, err := NewTextValue("Runs on shared Kubernetes cluster")
	require.NoError(t, err)
	number, err := NewNumberValue(42.5)
	require.NoError(t, err)
	date, err := NewDateValue("2026-03-01")
	require.NoError(t, err)
	link, err := NewLinkValue("MSA", "https://contracts.example.com")
	require.NoError(t, err)
	selection, err := NewSelectionValue(uuid.New().String())
	require.NoError(t, err)
	contact, err := NewContactPerson(ContactPersonParams{Name: "A. Larsen", Email: "al@ext.example", Company: "Ext ApS"})
	require.NoError(t, err)
	return []FieldValue{text, number, date, link, selection, contact}
}

func TestValueEnvelope_RoundTripsEveryFieldValueKind(t *testing.T) {
	for _, value := range mustFieldValues(t) {
		t.Run(value.FieldTypeValue(), func(t *testing.T) {
			envelope, err := NewValueEnvelope(value)
			require.NoError(t, err)
			assert.Equal(t, value.FieldTypeValue(), envelope.Type)
			assert.Equal(t, 1, envelope.Version)

			decoded, err := FieldValueFromEnvelope(envelope)
			require.NoError(t, err)
			assert.True(t, decoded.Equals(value))
		})
	}
}

func TestValueEnvelope_SurvivesJSONSerialization(t *testing.T) {
	link, err := NewLinkValue("MSA", "https://contracts.example.com")
	require.NoError(t, err)
	envelope, err := NewValueEnvelope(link)
	require.NoError(t, err)

	raw, err := json.Marshal(envelope)
	require.NoError(t, err)

	var restored ValueEnvelope
	require.NoError(t, json.Unmarshal(raw, &restored))

	decoded, err := FieldValueFromEnvelope(restored)
	require.NoError(t, err)
	assert.True(t, decoded.Equals(link))
}

func TestFieldValueFromEnvelope_RejectsUnknownType(t *testing.T) {
	_, err := FieldValueFromEnvelope(ValueEnvelope{Type: "geo", Version: 1, Value: json.RawMessage(`{}`)})
	assert.ErrorIs(t, err, ErrUnknownValueType)
}

func TestFieldValueFromEnvelope_RejectsUnsupportedVersion(t *testing.T) {
	for _, version := range []int{0, 2} {
		_, err := FieldValueFromEnvelope(ValueEnvelope{Type: "text", Version: version, Value: json.RawMessage(`"x"`)})
		assert.ErrorIs(t, err, ErrUnsupportedValueVersion)
	}
}

func TestFieldValueFromEnvelope_RejectsInvalidPayloads(t *testing.T) {
	cases := []struct {
		name     string
		envelope ValueEnvelope
	}{
		{"whitespace text", ValueEnvelope{Type: "text", Version: 1, Value: json.RawMessage(`"   "`)}},
		{"non-iso date", ValueEnvelope{Type: "date", Version: 1, Value: json.RawMessage(`"March 1st"`)}},
		{"relative link", ValueEnvelope{Type: "link", Version: 1, Value: json.RawMessage(`{"label":"MSA","url":"/x"}`)}},
		{"invalid contact email", ValueEnvelope{Type: "contact-person", Version: 1, Value: json.RawMessage(`{"name":"A","email":"not-an-email"}`)}},
		{"malformed payload", ValueEnvelope{Type: "number", Version: 1, Value: json.RawMessage(`"NaN"`)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := FieldValueFromEnvelope(tc.envelope)
			assert.Error(t, err)
		})
	}
}
