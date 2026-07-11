package queries_test

import (
	"context"
	"fmt"
	"testing"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getApplicationOnePager(t *testing.T, config *readmodels.ConfigurationRecord, snapshotFields map[string]ports.BuiltInFieldValue, facts []readmodels.FactRecord) (*queries.OnePager, error) {
	t.Helper()
	subjects := &countingSubjectSource{snapshot: snapshotNamed("App", snapshotFields)}
	query := queries.NewOnePagerQuery(buildDeps(depsParams{
		subjectType: "application",
		subjects:    subjects,
		configs:     &countingConfigSource{record: config},
		facts:       &countingFactsSource{records: facts},
	}))
	return query.Get(context.Background(), mustSubjectType(t, "application"), "app-1")
}

func TestGet_InterleavesBuiltInAndCustomFieldsInConfiguredOrder(t *testing.T) {
	config := &readmodels.ConfigurationRecord{
		SubjectType: "application",
		Document: readmodels.ConfigurationDocument{
			CustomFields: []readmodels.CustomFieldRecord{
				{ID: "contract-link", Name: "Contract link", Type: "link", Active: true},
			},
			DisplayOrder: []readmodels.FieldRefRecord{
				{Kind: "builtIn", ID: "description"},
				{Kind: "custom", ID: "contract-link"},
				{Kind: "builtIn", ID: "experts"},
			},
		},
	}

	result, err := getApplicationOnePager(t, config, map[string]ports.BuiltInFieldValue{
		"description": ports.TextValue{Text: "Handles payments"},
	}, nil)

	require.NoError(t, err)
	require.Len(t, result.Fields, 3)
	require.NotNil(t, result.Fields[0].BuiltIn)
	assert.Equal(t, "description", result.Fields[0].BuiltIn.ID)
	require.NotNil(t, result.Fields[1].Custom)
	assert.Equal(t, "contract-link", result.Fields[1].Custom.FieldID)
	require.NotNil(t, result.Fields[2].BuiltIn)
	assert.Equal(t, "experts", result.Fields[2].BuiltIn.ID)
}

func TestGet_BuiltInFieldResolvesLabelAndValueFromCatalogAndSnapshot(t *testing.T) {
	config := &readmodels.ConfigurationRecord{
		Document: readmodels.ConfigurationDocument{
			DisplayOrder: []readmodels.FieldRefRecord{{Kind: "builtIn", ID: "description"}},
		},
	}

	result, err := getApplicationOnePager(t, config, map[string]ports.BuiltInFieldValue{
		"description": ports.TextValue{Text: "Handles payments"},
	}, nil)

	require.NoError(t, err)
	require.Len(t, result.Fields, 1)
	require.NotNil(t, result.Fields[0].BuiltIn)
	assert.Equal(t, "description", result.Fields[0].BuiltIn.ID)
	assert.Equal(t, "Description", result.Fields[0].BuiltIn.Label)
	assert.Equal(t, ports.TextValue{Text: "Handles payments"}, result.Fields[0].BuiltIn.Value)
}

func TestGet_DisplayOrderEntryEdgeCases(t *testing.T) {
	cases := []struct {
		name       string
		ref        readmodels.FieldRefRecord
		wantFields int
	}{
		{"built-in with no snapshot value still renders", readmodels.FieldRefRecord{Kind: "builtIn", ID: "description"}, 1},
		{"unknown catalog entry is skipped", readmodels.FieldRefRecord{Kind: "builtIn", ID: "unknown-entry"}, 0},
		{"custom ref not found in configuration is skipped", readmodels.FieldRefRecord{Kind: "custom", ID: "ghost-field"}, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := &readmodels.ConfigurationRecord{
				Document: readmodels.ConfigurationDocument{DisplayOrder: []readmodels.FieldRefRecord{tc.ref}},
			}

			result, err := getApplicationOnePager(t, config, nil, nil)

			require.NoError(t, err)
			assert.Len(t, result.Fields, tc.wantFields)
			if tc.wantFields == 1 {
				assert.Nil(t, result.Fields[0].BuiltIn.Value)
			}
		})
	}
}

func TestGet_SkipsRetiredCustomFieldEvenWithRecordedValue(t *testing.T) {
	config := &readmodels.ConfigurationRecord{
		Document: readmodels.ConfigurationDocument{
			CustomFields: []readmodels.CustomFieldRecord{{ID: "old-field", Name: "Old", Type: "text", Active: false}},
			DisplayOrder: []readmodels.FieldRefRecord{{Kind: "custom", ID: "old-field"}},
		},
	}

	result, err := getApplicationOnePager(t, config, nil, []readmodels.FactRecord{
		{FieldID: "old-field", DisplayText: "value"},
	})

	require.NoError(t, err)
	assert.Empty(t, result.Fields)
}

func TestGet_EmptyOptionalCustomFieldIsPresentWithNilValue(t *testing.T) {
	config := &readmodels.ConfigurationRecord{
		Document: readmodels.ConfigurationDocument{
			CustomFields: []readmodels.CustomFieldRecord{{ID: "notes", Name: "Notes", Type: "text", HelpText: "Optional notes", Active: true}},
			DisplayOrder: []readmodels.FieldRefRecord{{Kind: "custom", ID: "notes"}},
		},
	}

	result, err := getApplicationOnePager(t, config, nil, nil)

	require.NoError(t, err)
	require.Len(t, result.Fields, 1)
	require.NotNil(t, result.Fields[0].Custom)
	assert.Equal(t, "notes", result.Fields[0].Custom.FieldID)
	assert.Equal(t, "Notes", result.Fields[0].Custom.Name)
	assert.Equal(t, "text", result.Fields[0].Custom.FieldType)
	assert.Equal(t, "Optional notes", result.Fields[0].Custom.HelpText)
	assert.Nil(t, result.Fields[0].Custom.Value)
	assert.Equal(t, "", result.Fields[0].Custom.DisplayText)
	assert.False(t, result.Fields[0].Custom.RetiredOption)
}

func TestGet_CustomFieldWithRecordedValueRendersValueAndDisplayText(t *testing.T) {
	envelope := envelopeOf(t, "text", `"Runs on shared cluster"`)
	config := &readmodels.ConfigurationRecord{
		Document: readmodels.ConfigurationDocument{
			CustomFields: []readmodels.CustomFieldRecord{{ID: "hosting-notes", Name: "Hosting notes", Type: "text", Active: true}},
			DisplayOrder: []readmodels.FieldRefRecord{{Kind: "custom", ID: "hosting-notes"}},
		},
	}

	result, err := getApplicationOnePager(t, config, nil, []readmodels.FactRecord{
		{FieldID: "hosting-notes", Value: &envelope, DisplayText: "Runs on shared cluster"},
	})

	require.NoError(t, err)
	require.Len(t, result.Fields, 1)
	assert.Equal(t, &envelope, result.Fields[0].Custom.Value)
	assert.Equal(t, "Runs on shared cluster", result.Fields[0].Custom.DisplayText)
}

func TestGet_SelectionFieldDisplayText(t *testing.T) {
	knownOptionID := "9f0d5e69-0000-0000-0000-00000000000d"

	cases := []struct {
		name                string
		options             []readmodels.OptionRecord
		selectedOptionID    string
		recordedDisplayText string
		wantDisplayText     string
		wantRetired         bool
	}{
		{
			name:                "active option resolves to its label",
			options:             []readmodels.OptionRecord{{ID: "9f0d5e69-0000-0000-0000-00000000000a", Label: "Cloud", Active: true}},
			selectedOptionID:    "9f0d5e69-0000-0000-0000-00000000000a",
			recordedDisplayText: "9f0d5e69-0000-0000-0000-00000000000a",
			wantDisplayText:     "Cloud",
			wantRetired:         false,
		},
		{
			name:                "retired option resolves to its label and flags retired",
			options:             []readmodels.OptionRecord{{ID: "9f0d5e69-0000-0000-0000-00000000000b", Label: "On-prem", Active: false}},
			selectedOptionID:    "9f0d5e69-0000-0000-0000-00000000000b",
			recordedDisplayText: "9f0d5e69-0000-0000-0000-00000000000b",
			wantDisplayText:     "On-prem",
			wantRetired:         true,
		},
		{
			name:                "unknown option falls back to the recorded display text",
			options:             []readmodels.OptionRecord{{ID: knownOptionID, Label: "Cloud", Active: true}},
			selectedOptionID:    "9f0d5e69-0000-0000-0000-00000000000c",
			recordedDisplayText: "stale display text",
			wantDisplayText:     "stale display text",
			wantRetired:         false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			envelope := envelopeOf(t, "selection", fmt.Sprintf(`{"optionId":%q}`, tc.selectedOptionID))
			config := &readmodels.ConfigurationRecord{
				Document: readmodels.ConfigurationDocument{
					CustomFields: []readmodels.CustomFieldRecord{{
						ID: "hosting", Name: "Hosting model", Type: "selection", Active: true,
						Options: tc.options,
					}},
					DisplayOrder: []readmodels.FieldRefRecord{{Kind: "custom", ID: "hosting"}},
				},
			}

			result, err := getApplicationOnePager(t, config, nil, []readmodels.FactRecord{
				{FieldID: "hosting", Value: &envelope, DisplayText: tc.recordedDisplayText},
			})

			require.NoError(t, err)
			require.Len(t, result.Fields, 1)
			assert.Equal(t, tc.wantDisplayText, result.Fields[0].Custom.DisplayText)
			assert.Equal(t, tc.wantRetired, result.Fields[0].Custom.RetiredOption)
		})
	}
}

func TestGet_FactsForFieldsOutsideDisplayOrderDoNotAppear(t *testing.T) {
	config := &readmodels.ConfigurationRecord{
		Document: readmodels.ConfigurationDocument{
			CustomFields: []readmodels.CustomFieldRecord{{ID: "notes", Name: "Notes", Type: "text", Active: true}},
			DisplayOrder: []readmodels.FieldRefRecord{{Kind: "custom", ID: "notes"}},
		},
	}

	result, err := getApplicationOnePager(t, config, nil, []readmodels.FactRecord{
		{FieldID: "notes", DisplayText: "kept"},
		{FieldID: "stray-field", DisplayText: "should not appear"},
	})

	require.NoError(t, err)
	require.Len(t, result.Fields, 1)
	assert.Equal(t, "kept", result.Fields[0].Custom.DisplayText)
}
