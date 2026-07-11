package readmodels_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func envelopeOf(valueType, rawValue string) valueobjects.ValueEnvelope {
	return valueobjects.ValueEnvelope{Type: valueType, Version: valueobjects.ValueEnvelopeVersion, Value: json.RawMessage(rawValue)}
}

func selectionEnvelope(optionID string) *valueobjects.ValueEnvelope {
	value := envelopeOf("selection", fmt.Sprintf(`{"optionId":%q}`, optionID))
	return &value
}

func TestCustomFieldRecord_SelectedOptionResolution(t *testing.T) {
	activeOptionID := uuid.New().String()
	retiredOptionID := uuid.New().String()
	selectionField := readmodels.CustomFieldRecord{
		ID: "hosting", Type: "selection", Active: true,
		Options: []readmodels.OptionRecord{
			{ID: activeOptionID, Label: "Cloud", Active: true},
			{ID: retiredOptionID, Label: "On-prem", Active: false},
		},
	}
	nonSelectionField := readmodels.CustomFieldRecord{ID: "notes", Type: "text"}
	nonSelectionValue := envelopeOf("text", `"just text"`)
	undecodableValue := envelopeOf("text", `"just text"`)
	undecodableValue.Version = 99

	cases := []struct {
		name         string
		field        readmodels.CustomFieldRecord
		value        *valueobjects.ValueEnvelope
		wantRetired  bool
		wantLabel    string
		wantLabelled bool
	}{
		{
			name:         "active option is not retired and resolves its label",
			field:        selectionField,
			value:        selectionEnvelope(activeOptionID),
			wantRetired:  false,
			wantLabel:    "Cloud",
			wantLabelled: true,
		},
		{
			name:         "retired option is flagged retired and still resolves its label",
			field:        selectionField,
			value:        selectionEnvelope(retiredOptionID),
			wantRetired:  true,
			wantLabel:    "On-prem",
			wantLabelled: true,
		},
		{
			name:         "unknown option id is neither retired nor labelled",
			field:        selectionField,
			value:        selectionEnvelope(uuid.New().String()),
			wantRetired:  false,
			wantLabelled: false,
		},
		{
			name:         "non-selection value is neither retired nor labelled",
			field:        nonSelectionField,
			value:        &nonSelectionValue,
			wantRetired:  false,
			wantLabelled: false,
		},
		{
			name:         "undecodable envelope is neither retired nor labelled",
			field:        nonSelectionField,
			value:        &undecodableValue,
			wantRetired:  false,
			wantLabelled: false,
		},
		{
			name:         "nil value is neither retired nor labelled",
			field:        selectionField,
			value:        nil,
			wantRetired:  false,
			wantLabelled: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantRetired, tc.field.RetiredOptionReferenced(tc.value))

			label, labelled := tc.field.SelectionOptionLabel(tc.value)
			assert.Equal(t, tc.wantLabelled, labelled)
			assert.Equal(t, tc.wantLabel, label)
		})
	}
}

func TestConfigurationDocument_CustomField_FindsByID(t *testing.T) {
	document := readmodels.ConfigurationDocument{
		CustomFields: []readmodels.CustomFieldRecord{
			{ID: "hosting", Name: "Hosting model"},
			{ID: "notes", Name: "Notes"},
		},
	}

	field, found := document.CustomField("notes")

	assert.True(t, found)
	assert.Equal(t, "Notes", field.Name)
}

func TestConfigurationDocument_CustomField_NotFound(t *testing.T) {
	document := readmodels.ConfigurationDocument{
		CustomFields: []readmodels.CustomFieldRecord{{ID: "hosting", Name: "Hosting model"}},
	}

	_, found := document.CustomField("missing")

	assert.False(t, found)
}
