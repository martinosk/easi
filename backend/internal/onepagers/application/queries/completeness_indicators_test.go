package queries_test

import (
	"context"
	"testing"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noBuiltInSources() map[string]ports.BuiltInFieldSource {
	return map[string]ports.BuiltInFieldSource{}
}

func builtInSources(source ports.BuiltInFieldSource) map[string]ports.BuiltInFieldSource {
	return map[string]ports.BuiltInFieldSource{"application": source}
}

func requiredBuiltInIndicatorConfig(customFields ...indicatorSpec) *readmodels.ConfigurationRecord {
	config := indicatorConfigWith(customFields...)
	config.SubjectType = "application"
	config.Document.BuiltInFields = []readmodels.FieldRequirementRecord{{ID: "experts", Required: true}}
	config.Document.DisplayOrder = append(config.Document.DisplayOrder, readmodels.FieldRefRecord{Kind: "builtIn", ID: "experts"})
	return config
}

type stubConfigurationSource struct {
	calls  int
	record *readmodels.ConfigurationRecord
	err    error
}

func (s *stubConfigurationSource) GetBySubjectType(_ context.Context, _ string) (*readmodels.ConfigurationRecord, error) {
	s.calls++
	return s.record, s.err
}

func (s *stubConfigurationSource) ForSubjectType(_ context.Context, _ string) (readmodels.CustomFieldDefinitions, error) {
	if s.record == nil {
		return nil, s.err
	}
	included := map[string]bool{}
	for _, id := range s.record.Document.IncludedCustomFieldIDs() {
		included[id] = true
	}
	var definitions readmodels.CustomFieldDefinitions
	for _, field := range s.record.Document.CustomFields {
		definitions = append(definitions, readmodels.CustomFieldRecord{ID: field.ID, Name: field.ID, Type: "text", Active: included[field.ID]})
	}
	return definitions, nil
}

type stubFilledCountsSource struct {
	calls  int
	gotIDs []string
	counts map[string]int
	err    error
}

func (s *stubFilledCountsSource) FilledFieldCounts(_ context.Context, _ string, subjectIDs, _ []string) (map[string]int, error) {
	s.calls++
	s.gotIDs = subjectIDs
	return s.counts, s.err
}

type indicatorSpec struct {
	id       string
	required bool
	included bool
}

func indicatorField(id string, required, included bool) indicatorSpec {
	return indicatorSpec{id: id, required: required, included: included}
}

func indicatorConfigWith(fields ...indicatorSpec) *readmodels.ConfigurationRecord {
	document := readmodels.ConfigurationDocument{}
	for _, field := range fields {
		document.CustomFields = append(document.CustomFields, readmodels.FieldRequirementRecord{ID: field.id, Required: field.required})
		if field.included {
			document.DisplayOrder = append(document.DisplayOrder, readmodels.FieldRefRecord{Kind: "custom", ID: field.id})
		}
	}
	return &readmodels.ConfigurationRecord{Document: document}
}

func TestForSubjects_IndicatorNotApplicable_SkipsFactsQuery(t *testing.T) {
	cases := []struct {
		name   string
		config *readmodels.ConfigurationRecord
	}{
		{"no configuration", nil},
		{"only optional fields", indicatorConfigWith(indicatorField("notes", false, true))},
		{"only retired required fields", indicatorConfigWith(indicatorField("contract-link", true, false))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			facts := &stubFilledCountsSource{}
			sut := queries.NewCompletenessIndicators(
				&stubConfigurationSource{record: tc.config},
				&stubConfigurationSource{record: tc.config}, facts, noBuiltInSources())

			result, present, err := sut.ForSubjects(context.Background(), "application", []string{"app-1"})

			require.NoError(t, err)
			assert.False(t, present)
			assert.Nil(t, result)
			assert.Equal(t, 0, facts.calls)
		})
	}
}

func TestForSubjects_IndicatorValues(t *testing.T) {
	twoRequired := indicatorConfigWith(indicatorField("field-a", true, true), indicatorField("field-b", true, true))
	oneRequired := indicatorConfigWith(indicatorField("field-a", true, true))

	cases := []struct {
		name           string
		config         *readmodels.ConfigurationRecord
		counts         map[string]int
		subjectIDs     []string
		want           map[string]bool
		wantFactsCalls int
	}{
		{
			name:           "all required fields filled is complete",
			config:         twoRequired,
			counts:         map[string]int{"app-1": 2},
			subjectIDs:     []string{"app-1"},
			want:           map[string]bool{"app-1": true},
			wantFactsCalls: 1,
		},
		{
			name:           "partially filled is incomplete",
			config:         twoRequired,
			counts:         map[string]int{"app-1": 1},
			subjectIDs:     []string{"app-1"},
			want:           map[string]bool{"app-1": false},
			wantFactsCalls: 1,
		},
		{
			name:           "subject absent from counts is incomplete",
			config:         oneRequired,
			counts:         map[string]int{},
			subjectIDs:     []string{"app-1"},
			want:           map[string]bool{"app-1": false},
			wantFactsCalls: 1,
		},
		{
			name:           "every requested id is present in the result",
			config:         oneRequired,
			counts:         map[string]int{"app-1": 1},
			subjectIDs:     []string{"app-1", "app-2", "app-3"},
			want:           map[string]bool{"app-1": true, "app-2": false, "app-3": false},
			wantFactsCalls: 1,
		},
		{
			name:           "empty subject ids returns empty map without querying facts",
			config:         oneRequired,
			counts:         map[string]int{},
			subjectIDs:     []string{},
			want:           map[string]bool{},
			wantFactsCalls: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			facts := &stubFilledCountsSource{counts: tc.counts}
			sut := queries.NewCompletenessIndicators(
				&stubConfigurationSource{record: tc.config},
				&stubConfigurationSource{record: tc.config}, facts, noBuiltInSources())

			result, present, err := sut.ForSubjects(context.Background(), "application", tc.subjectIDs)

			require.NoError(t, err)
			assert.True(t, present)
			assert.Equal(t, tc.want, result)
			assert.Equal(t, tc.wantFactsCalls, facts.calls)
			if tc.wantFactsCalls > 0 {
				assert.Equal(t, tc.subjectIDs, facts.gotIDs)
			}
		})
	}
}

func TestForSubjects_ErrorPropagation(t *testing.T) {
	requiredConfig := indicatorConfigWith(indicatorField("field-a", true, true))

	cases := []struct {
		name    string
		configs *stubConfigurationSource
		facts   *stubFilledCountsSource
	}{
		{
			name:    "configuration load error",
			configs: &stubConfigurationSource{err: assert.AnError},
			facts:   &stubFilledCountsSource{},
		},
		{
			name:    "facts query error",
			configs: &stubConfigurationSource{record: requiredConfig},
			facts:   &stubFilledCountsSource{err: assert.AnError},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sut := queries.NewCompletenessIndicators(tc.configs, tc.configs, tc.facts, noBuiltInSources())

			result, present, err := sut.ForSubjects(context.Background(), "application", []string{"app-1"})

			require.ErrorIs(t, err, assert.AnError)
			assert.False(t, present)
			assert.Nil(t, result)
		})
	}
}

func TestForSubjects_RequiredBuiltInIncludedInIndicator(t *testing.T) {
	cases := []struct {
		name       string
		config     *readmodels.ConfigurationRecord
		counts     map[string]int
		filled     map[string]map[string]bool
		subjectIDs []string
		want       map[string]bool
	}{
		{
			name:       "built-in only, filled subject is complete and empty subject is incomplete",
			config:     requiredBuiltInIndicatorConfig(),
			filled:     map[string]map[string]bool{"app-1": {"experts": true}, "app-2": {"experts": false}},
			subjectIDs: []string{"app-1", "app-2"},
			want:       map[string]bool{"app-1": true, "app-2": false},
		},
		{
			name:       "custom and built-in both required needs both filled",
			config:     requiredBuiltInIndicatorConfig(indicatorField("contract-link", true, true)),
			counts:     map[string]int{"app-1": 1, "app-2": 1},
			filled:     map[string]map[string]bool{"app-1": {"experts": true}, "app-2": {"experts": false}},
			subjectIDs: []string{"app-1", "app-2"},
			want:       map[string]bool{"app-1": true, "app-2": false},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			facts := &stubFilledCountsSource{counts: tc.counts}
			source := &countingSubjectSource{filled: tc.filled}
			sut := queries.NewCompletenessIndicators(
				&stubConfigurationSource{record: tc.config},
				&stubConfigurationSource{record: tc.config}, facts, builtInSources(source))

			result, present, err := sut.ForSubjects(context.Background(), "application", tc.subjectIDs)

			require.NoError(t, err)
			assert.True(t, present)
			assert.Equal(t, tc.want, result)
			assert.Equal(t, 1, source.filledCalls)
			assert.Equal(t, tc.subjectIDs, source.gotFilledIDs)
			assert.Equal(t, []string{"experts"}, source.gotEntryIDs)
		})
	}
}

func TestForSubjects_BuiltInFillErrorPropagates(t *testing.T) {
	source := &countingSubjectSource{filledErr: assert.AnError}
	sut := queries.NewCompletenessIndicators(
		&stubConfigurationSource{record: requiredBuiltInIndicatorConfig()},
		&stubConfigurationSource{record: requiredBuiltInIndicatorConfig()},
		&stubFilledCountsSource{},
		builtInSources(source),
	)

	_, present, err := sut.ForSubjects(context.Background(), "application", []string{"app-1"})

	require.ErrorIs(t, err, assert.AnError)
	assert.False(t, present)
}
