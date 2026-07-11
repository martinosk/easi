package queries_test

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet_CallsEachCollaboratorExactlyOnce(t *testing.T) {
	subjects := &countingSubjectSource{snapshot: snapshotNamed("Payments", nil)}
	configs := &countingConfigSource{record: &readmodels.ConfigurationRecord{SubjectType: "capability"}}
	facts := &countingFactsSource{}
	maturity := &countingMaturitySource{}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: configs, facts: facts, maturity: maturity}))

	_, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "subject-1")

	require.NoError(t, err)
	assert.Equal(t, 1, configs.calls)
	assert.Equal(t, 1, subjects.calls)
	assert.Equal(t, 1, facts.calls)
	assert.Equal(t, 0, maturity.calls)
	assert.Equal(t, "capability", configs.gotSubjectType)
	assert.Equal(t, readmodels.SubjectKey{SubjectType: "capability", SubjectID: "subject-1"}, facts.gotSubject)
}

func TestGet_ReturnsErrSubjectNotFoundWhenFetchSubjectReturnsNil(t *testing.T) {
	subjects := &countingSubjectSource{snapshot: nil}
	configs := &countingConfigSource{}
	facts := &countingFactsSource{}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: configs, facts: facts}))

	_, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "missing")

	assert.ErrorIs(t, err, queries.ErrSubjectNotFound)
	assert.Equal(t, 1, subjects.calls)
	assert.Equal(t, 0, configs.calls)
	assert.Equal(t, 0, facts.calls)
}

func TestGet_ReturnsErrSubjectNotFoundWhenSubjectTypeMissingFromMap(t *testing.T) {
	configs := &countingConfigSource{}
	facts := &countingFactsSource{}

	query := queries.NewOnePagerQuery(queries.OnePagerQueryDeps{
		Configurations: configs,
		Facts:          facts,
		Subjects:       map[string]ports.BuiltInFieldSource{},
	})

	_, err := query.Get(context.Background(), mustSubjectType(t, "vendor"), "subject-1")

	assert.ErrorIs(t, err, queries.ErrSubjectNotFound)
	assert.Equal(t, 0, configs.calls)
	assert.Equal(t, 0, facts.calls)
}

func TestGet_SetsSubjectHeaderFields(t *testing.T) {
	subjects := &countingSubjectSource{snapshot: snapshotNamed("Payments Capability", nil)}
	configs := &countingConfigSource{record: &readmodels.ConfigurationRecord{SubjectType: "capability"}}
	facts := &countingFactsSource{}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: configs, facts: facts}))

	result, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "cap-42")

	require.NoError(t, err)
	assert.Equal(t, "capability", result.SubjectType)
	assert.Equal(t, "cap-42", result.SubjectID)
	assert.Equal(t, "Payments Capability", result.SubjectName)
}

func TestGet_PropagatesFetchSubjectError(t *testing.T) {
	wantErr := errors.New("boom")
	subjects := &countingSubjectSource{err: wantErr}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: &countingConfigSource{}, facts: &countingFactsSource{}}))

	_, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "subject-1")

	assert.ErrorIs(t, err, wantErr)
}

func TestGet_PropagatesConfigurationsError(t *testing.T) {
	wantErr := errors.New("boom")
	subjects := &countingSubjectSource{snapshot: snapshotNamed("Payments", nil)}
	configs := &countingConfigSource{err: wantErr}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: configs, facts: &countingFactsSource{}}))

	_, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "subject-1")

	assert.ErrorIs(t, err, wantErr)
}

func TestGet_PropagatesFactsError(t *testing.T) {
	wantErr := errors.New("boom")
	subjects := &countingSubjectSource{snapshot: snapshotNamed("Payments", nil)}
	facts := &countingFactsSource{err: wantErr}

	query := queries.NewOnePagerQuery(buildDeps(depsParams{subjectType: "capability", subjects: subjects, configs: &countingConfigSource{}, facts: facts}))

	_, err := query.Get(context.Background(), mustSubjectType(t, "capability"), "subject-1")

	assert.ErrorIs(t, err, wantErr)
}
