package queries_test

import (
	"testing"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func singleCustomFieldConfig(field readmodels.CustomFieldRecord) *readmodels.ConfigurationRecord {
	return &readmodels.ConfigurationRecord{
		Document: readmodels.ConfigurationDocument{
			CustomFields: []readmodels.CustomFieldRecord{field},
			DisplayOrder: []readmodels.FieldRefRecord{{Kind: "custom", ID: field.ID}},
		},
	}
}

func missingFieldIDs(result *queries.OnePager) []string {
	ids := make([]string, len(result.Completeness.MissingFields))
	for i, field := range result.Completeness.MissingFields {
		ids[i] = field.FieldID
	}
	return ids
}

func customFieldsConfig(fields ...readmodels.CustomFieldRecord) *readmodels.ConfigurationRecord {
	order := make([]readmodels.FieldRefRecord, len(fields))
	for i, field := range fields {
		order[i] = readmodels.FieldRefRecord{Kind: "custom", ID: field.ID}
	}
	return &readmodels.ConfigurationRecord{
		Document: readmodels.ConfigurationDocument{CustomFields: fields, DisplayOrder: order},
	}
}

func TestGet_CompletenessSingleRequiredFieldScenarios(t *testing.T) {
	envelope := envelopeOf(t, "text", `"value"`)
	requiredActiveField := readmodels.CustomFieldRecord{ID: "contract-link", Name: "Contract link", Type: "link", Required: true, Active: true}
	requiredRetiredField := readmodels.CustomFieldRecord{ID: "contract-link", Name: "Contract link", Type: "link", Required: true, Active: false}

	cases := []struct {
		name              string
		config            *readmodels.ConfigurationRecord
		facts             []readmodels.FactRecord
		wantRequiredCount int
		wantFilledCount   int
		wantMissingIDs    []string
	}{
		{
			name:              "retired required field is excluded entirely",
			config:            singleCustomFieldConfig(requiredRetiredField),
			wantRequiredCount: 0,
			wantFilledCount:   0,
		},
		{
			name:              "reactivated required field is counted again and incomplete",
			config:            singleCustomFieldConfig(requiredActiveField),
			wantRequiredCount: 1,
			wantFilledCount:   0,
			wantMissingIDs:    []string{"contract-link"},
		},
		{
			name:              "fact with nil value counts as missing",
			config:            singleCustomFieldConfig(requiredActiveField),
			facts:             []readmodels.FactRecord{{FieldID: "contract-link", Value: nil}},
			wantRequiredCount: 1,
			wantFilledCount:   0,
			wantMissingIDs:    []string{"contract-link"},
		},
		{
			name:              "fact with a value counts as filled",
			config:            singleCustomFieldConfig(requiredActiveField),
			facts:             []readmodels.FactRecord{{FieldID: "contract-link", Value: &envelope}},
			wantRequiredCount: 1,
			wantFilledCount:   1,
		},
		{
			name:              "no configuration defaults to zero",
			config:            nil,
			wantRequiredCount: 0,
			wantFilledCount:   0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getApplicationOnePager(t, tc.config, nil, tc.facts)

			require.NoError(t, err)
			assert.Equal(t, tc.wantRequiredCount, result.Completeness.RequiredCount)
			assert.Equal(t, tc.wantFilledCount, result.Completeness.FilledCount)
			assert.ElementsMatch(t, tc.wantMissingIDs, missingFieldIDs(result))
		})
	}
}

func TestGet_CompletenessMultipleCustomFieldScenarios(t *testing.T) {
	envelope := envelopeOf(t, "text", `"value"`)

	cases := []struct {
		name              string
		fields            []readmodels.CustomFieldRecord
		facts             []readmodels.FactRecord
		wantRequiredCount int
		wantFilledCount   int
		wantMissing       []queries.MissingField
	}{
		{
			name: "both required fields filled",
			fields: []readmodels.CustomFieldRecord{
				{ID: "field-a", Name: "Field A", Type: "text", Required: true, Active: true},
				{ID: "field-b", Name: "Field B", Type: "text", Required: true, Active: true},
			},
			facts: []readmodels.FactRecord{
				{FieldID: "field-a", Value: &envelope},
				{FieldID: "field-b", Value: &envelope},
			},
			wantRequiredCount: 2,
			wantFilledCount:   2,
		},
		{
			name: "missing required field is named",
			fields: []readmodels.CustomFieldRecord{
				{ID: "contract-link", Name: "Contract link", Type: "link", Required: true, Active: true},
				{ID: "contact-person", Name: "Contact person", Type: "text", Required: true, Active: true},
			},
			facts: []readmodels.FactRecord{
				{FieldID: "contact-person", Value: &envelope},
			},
			wantRequiredCount: 2,
			wantFilledCount:   1,
			wantMissing:       []queries.MissingField{{FieldID: "contract-link", Name: "Contract link"}},
		},
		{
			name: "optional field without value is not counted",
			fields: []readmodels.CustomFieldRecord{
				{ID: "required-field", Name: "Required field", Type: "text", Required: true, Active: true},
				{ID: "notes", Name: "Notes", Type: "text", Required: false, Active: true},
			},
			facts: []readmodels.FactRecord{
				{FieldID: "required-field", Value: &envelope},
			},
			wantRequiredCount: 1,
			wantFilledCount:   1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getApplicationOnePager(t, customFieldsConfig(tc.fields...), nil, tc.facts)

			require.NoError(t, err)
			assert.Equal(t, tc.wantRequiredCount, result.Completeness.RequiredCount)
			assert.Equal(t, tc.wantFilledCount, result.Completeness.FilledCount)
			assert.ElementsMatch(t, tc.wantMissing, result.Completeness.MissingFields)
		})
	}
}

func TestGet_CompletenessIgnoresNonRequiredBuiltInFields(t *testing.T) {
	config := &readmodels.ConfigurationRecord{
		SubjectType: "application",
		Document: readmodels.ConfigurationDocument{
			DisplayOrder: []readmodels.FieldRefRecord{
				{Kind: "builtIn", ID: "description"},
			},
		},
	}

	result, err := getApplicationOnePager(t, config, map[string]ports.BuiltInFieldValue{
		"description": ports.TextValue{Text: "Handles payments"},
	}, nil)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Completeness.RequiredCount)
	assert.Equal(t, 0, result.Completeness.FilledCount)
	assert.Empty(t, result.Completeness.MissingFields)
}

func requiredBuiltInConfig(entryID string) *readmodels.ConfigurationRecord {
	return &readmodels.ConfigurationRecord{
		SubjectType: "application",
		Document: readmodels.ConfigurationDocument{
			BuiltInFields: []readmodels.BuiltInFieldRecord{{ID: entryID, Required: true}},
			DisplayOrder:  []readmodels.FieldRefRecord{{Kind: "builtIn", ID: entryID}},
		},
	}
}

func TestGet_CompletenessRequiredBuiltInScenarios(t *testing.T) {
	cases := []struct {
		name              string
		entryID           string
		snapshot          map[string]ports.BuiltInFieldValue
		wantRequiredCount int
		wantFilledCount   int
		wantMissing       []queries.MissingField
	}{
		{
			name:              "required built-in with no value flags missing",
			entryID:           "experts",
			snapshot:          nil,
			wantRequiredCount: 1,
			wantFilledCount:   0,
			wantMissing:       []queries.MissingField{{FieldID: "experts", Name: "Experts"}},
		},
		{
			name:              "populated required built-in counts as filled",
			entryID:           "experts",
			snapshot:          map[string]ports.BuiltInFieldValue{"experts": ports.ExpertsValue{Experts: []ports.Expert{{Name: "Alice"}}}},
			wantRequiredCount: 1,
			wantFilledCount:   1,
		},
		{
			name:              "required experts with empty list counts as missing",
			entryID:           "experts",
			snapshot:          map[string]ports.BuiltInFieldValue{"experts": ports.ExpertsValue{Experts: nil}},
			wantRequiredCount: 1,
			wantFilledCount:   0,
			wantMissing:       []queries.MissingField{{FieldID: "experts", Name: "Experts"}},
		},
		{
			name:              "required text built-in with a value counts as filled",
			entryID:           "description",
			snapshot:          map[string]ports.BuiltInFieldValue{"description": ports.TextValue{Text: "Handles payments"}},
			wantRequiredCount: 1,
			wantFilledCount:   1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getApplicationOnePager(t, requiredBuiltInConfig(tc.entryID), tc.snapshot, nil)

			require.NoError(t, err)
			assert.Equal(t, tc.wantRequiredCount, result.Completeness.RequiredCount)
			assert.Equal(t, tc.wantFilledCount, result.Completeness.FilledCount)
			assert.ElementsMatch(t, tc.wantMissing, result.Completeness.MissingFields)
		})
	}
}

func TestGet_CompletenessCombinesRequiredCustomAndBuiltIn(t *testing.T) {
	envelope := envelopeOf(t, "text", `"value"`)
	config := &readmodels.ConfigurationRecord{
		SubjectType: "application",
		Document: readmodels.ConfigurationDocument{
			CustomFields:  []readmodels.CustomFieldRecord{{ID: "contract-link", Name: "Contract link", Type: "link", Required: true, Active: true}},
			BuiltInFields: []readmodels.BuiltInFieldRecord{{ID: "experts", Required: true}},
			DisplayOrder: []readmodels.FieldRefRecord{
				{Kind: "custom", ID: "contract-link"},
				{Kind: "builtIn", ID: "experts"},
			},
		},
	}

	result, err := getApplicationOnePager(t, config, nil, []readmodels.FactRecord{{FieldID: "contract-link", Value: &envelope}})

	require.NoError(t, err)
	assert.Equal(t, 2, result.Completeness.RequiredCount)
	assert.Equal(t, 1, result.Completeness.FilledCount)
	assert.Equal(t, []queries.MissingField{{FieldID: "experts", Name: "Experts"}}, result.Completeness.MissingFields)
}

func TestGet_CompletenessExcludedRequiredBuiltInDoesNotParticipate(t *testing.T) {
	config := &readmodels.ConfigurationRecord{
		SubjectType: "application",
		Document: readmodels.ConfigurationDocument{
			BuiltInFields: []readmodels.BuiltInFieldRecord{{ID: "experts", Required: true}},
			DisplayOrder:  []readmodels.FieldRefRecord{{Kind: "builtIn", ID: "description"}},
		},
	}

	result, err := getApplicationOnePager(t, config, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Completeness.RequiredCount)
	assert.Empty(t, result.Completeness.MissingFields)
}
