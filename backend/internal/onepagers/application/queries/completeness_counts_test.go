package queries_test

import (
	"context"
	"testing"

	"easi/backend/internal/onepagers/application/queries"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountsForSubjects_NoRequiredFields(t *testing.T) {
	facts := &stubFilledCountsSource{}
	sut := queries.NewCompletenessIndicators(
		&stubConfigurationSource{record: indicatorConfigWith(indicatorField("notes", false, true))},
		facts,
		noBuiltInSources(),
	)

	required, filled, err := sut.CountsForSubjects(context.Background(), "application", []string{"app-1", "app-2"})

	require.NoError(t, err)
	assert.Equal(t, 0, required)
	assert.Equal(t, map[string]int{"app-1": 0, "app-2": 0}, filled)
	assert.Equal(t, 0, facts.calls)
}

func TestCountsForSubjects_CustomAndBuiltIn(t *testing.T) {
	facts := &stubFilledCountsSource{counts: map[string]int{"app-1": 1, "app-2": 1}}
	source := &countingSubjectSource{filled: map[string]map[string]bool{
		"app-1": {"experts": true},
		"app-2": {"experts": false},
	}}
	sut := queries.NewCompletenessIndicators(
		&stubConfigurationSource{record: requiredBuiltInIndicatorConfig(indicatorField("contract-link", true, true))},
		facts,
		builtInSources(source),
	)

	required, filled, err := sut.CountsForSubjects(context.Background(), "application", []string{"app-1", "app-2"})

	require.NoError(t, err)
	assert.Equal(t, 2, required)
	assert.Equal(t, map[string]int{"app-1": 2, "app-2": 1}, filled)
}

func TestCountsForSubjects_EmptySubjects(t *testing.T) {
	facts := &stubFilledCountsSource{}
	sut := queries.NewCompletenessIndicators(
		&stubConfigurationSource{record: requiredBuiltInIndicatorConfig()},
		facts,
		builtInSources(&countingSubjectSource{}),
	)

	required, filled, err := sut.CountsForSubjects(context.Background(), "application", nil)

	require.NoError(t, err)
	assert.Equal(t, 1, required)
	assert.Equal(t, map[string]int{}, filled)
}

func TestCountsForSubjects_ErrorPropagates(t *testing.T) {
	sut := queries.NewCompletenessIndicators(
		&stubConfigurationSource{err: assert.AnError},
		&stubFilledCountsSource{},
		noBuiltInSources(),
	)

	_, _, err := sut.CountsForSubjects(context.Background(), "application", []string{"app-1"})

	require.ErrorIs(t, err, assert.AnError)
}
