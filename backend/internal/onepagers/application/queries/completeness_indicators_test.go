package queries_test

import (
	"context"
	"testing"

	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubConfigurationSource struct {
	calls  int
	record *readmodels.ConfigurationRecord
	err    error
}

func (s *stubConfigurationSource) GetBySubjectType(_ context.Context, _ string) (*readmodels.ConfigurationRecord, error) {
	s.calls++
	return s.record, s.err
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

func indicatorField(id string, required, active bool) readmodels.CustomFieldRecord {
	return readmodels.CustomFieldRecord{ID: id, Name: id, Type: "text", Required: required, Active: active}
}

func indicatorConfigWith(fields ...readmodels.CustomFieldRecord) *readmodels.ConfigurationRecord {
	return &readmodels.ConfigurationRecord{Document: readmodels.ConfigurationDocument{CustomFields: fields}}
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
			sut := queries.NewCompletenessIndicators(&stubConfigurationSource{record: tc.config}, facts)

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
			sut := queries.NewCompletenessIndicators(&stubConfigurationSource{record: tc.config}, facts)

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
			sut := queries.NewCompletenessIndicators(tc.configs, tc.facts)

			result, present, err := sut.ForSubjects(context.Background(), "application", []string{"app-1"})

			require.ErrorIs(t, err, assert.AnError)
			assert.False(t, present)
			assert.Nil(t, result)
		})
	}
}
